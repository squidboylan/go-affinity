package mesh

import (
	"fmt"
	"github.com/squidboylan/go-affinity/functional/utils"
	"github.com/project-receptor/receptor/pkg/backends"
	"github.com/project-receptor/receptor/pkg/netceptor"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	_ "sync"
	"time"
	"gopkg.in/yaml.v2"
	"strconv"
	"errors"
)

type Node struct {
	NetceptorInstance *netceptor.Netceptor
}

type Mesh struct {
	Nodes map[string]Node
	MeshDefinition YamlData
}

type YamlData struct {
	Nodes map[string] *struct {
		Connections map[string]float64
		Listen string
		Name string
		Stats_enable bool
		Stats_port string
	}
}

// Error handler that gets called for backend errors
func handleError(err error, fatal bool) {
	fmt.Printf("Error: %s\n", err)
	if fatal {
		os.Exit(1)
	}
}

func NewNode(name string) Node {
	n1 := netceptor.New(name, nil)
	return Node {
		NetceptorInstance: n1,
	}
}

func (n *Node) TCPListen(address string, cost float64) error {
	b1, err := backends.NewTCPListener(address, nil)
	if err != nil {
		return err
	}
	n.NetceptorInstance.RunBackend(b1, cost, handleError)
	return err
}

func (n *Node) TCPDial(address string, cost float64) error {
	b1, err := backends.NewTCPDialer(address, true, nil)
	if err != nil {
		return err
	}
	n.NetceptorInstance.RunBackend(b1, cost, handleError)
	return err
}

func (n *Node) ServiceListen(name string, function func(*netceptor.Listener)) (*netceptor.Listener, error) {
	l1, err := n.NetceptorInstance.Listen(name, nil)
	if err != nil {
		return nil, err
	}
	go function(l1)
	return l1, err
}

func (n *Node) ServiceDial(node string, servicename string, timeout int, function func()) (net.Conn, error) {
	for timeout > 0 {
		fmt.Printf("Dialing node1\n")
		c2, err := n.NetceptorInstance.Dial("node1", "echo", nil)
		if err != nil {
			fmt.Printf("Error dialing on Receptor network: %s\n", err)
			time.Sleep(1 * time.Second)
			continue
		}
		return c2, nil
	}
	return nil, fmt.Errorf("Timed out connecting to %s", node)
}

func NewMeshFromFile(filename string) Mesh {
	Nodes := make(map[string]Node)

	yamlDat, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("failed to read %s", filename)
		os.Exit(1)
	}

	MeshDefinition := YamlData {}

	yaml.Unmarshal(yamlDat, &MeshDefinition)

	for k := range MeshDefinition.Nodes {
		node := NewNode(MeshDefinition.Nodes[k].Name)
		if MeshDefinition.Nodes[k].Listen != "" {
			err := node.TCPListen(MeshDefinition.Nodes[k].Listen, 1.0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			
		} else {
			retries := 5
			for retries > 0 {
				port := utils.RandomPort()
				addrString := "127.0.0.1:" + strconv.Itoa(port)
				err := node.TCPListen(addrString, 1.0)
				if err == nil {
					MeshDefinition.Nodes[k].Listen = addrString
					break
				}
				retries -= 1
			}
			if retries == 0 {
				fmt.Println("Failed to conenct to a port after trying 5 times")
				os.Exit(1)
			}
		}
		Nodes[MeshDefinition.Nodes[k].Name] = node
	}
	for k := range MeshDefinition.Nodes {
		node := Nodes[MeshDefinition.Nodes[k].Name]
		for conn, cost := range MeshDefinition.Nodes[k].Connections {
			node.TCPDial(MeshDefinition.Nodes[conn].Listen, cost)
		}
	}
	return Mesh {
		Nodes,
		MeshDefinition,
	}
}

// This is broken and causes the thread to hang, dont use until
// netceptor.Shutdown is fixed
func (m *Mesh)Shutdown() {
	for _, node := range m.Nodes {
		node.NetceptorInstance.Shutdown()
	}
}

func (m *Mesh)CheckConnections() bool {
	for _, status := range m.Status() {
		actualConnections := map[string]float64{}
		for _, connection := range status.Connections {
			actualConnections[connection.NodeID] = connection.Cost
		}
		expectedConnections := map[string]float64{}
		for k, v := range m.MeshDefinition.Nodes[status.NodeID].Connections {
			expectedConnections[k] = v
		}
		for nodeID, node := range m.MeshDefinition.Nodes {
			if nodeID == status.NodeID {
				continue
			}
			for k, v := range node.Connections {
				if k == status.NodeID {
					expectedConnections[nodeID] = v
				}
			}
		}
		if reflect.DeepEqual(actualConnections, expectedConnections) {
			return true
		}
	}
	return false
}

func (m *Mesh)CheckKnownConnectionCosts() bool {
	meshStatus := m.Status()
	// If the mesh is empty we are done
	if len(meshStatus) == 0 {
		return true
	}

	knownConnectionCosts := meshStatus[0].KnownConnectionCosts
	for _, status := range m.Status() {
		if !reflect.DeepEqual(status.KnownConnectionCosts, knownConnectionCosts) {
			return false
		}
	}
	return true
}

func (m *Mesh)WaitForReady(timeout float64) error {
	connectionsReady := m.CheckConnections()
	for ;timeout > 0 && !connectionsReady; connectionsReady = m.CheckConnections() {
		time.Sleep(200 * time.Millisecond)
		timeout -= 200
	}
	if connectionsReady == false {
		return errors.New("Timed out while waiting for connections")
	}

	routesReady := m.CheckKnownConnectionCosts()
	for ;timeout > 0 && !routesReady; routesReady = m.CheckKnownConnectionCosts() {
		time.Sleep(200 * time.Millisecond)
		timeout -= 200
	}
	if routesReady == false {
		return errors.New("Timed out while waiting for routes to settle")
	}

	return nil
}

func (m *Mesh)Status() []netceptor.Status {
	out := []netceptor.Status{}
	for _, node := range m.Nodes {
		out = append(out, node.NetceptorInstance.Status())
	}
	return out
}
