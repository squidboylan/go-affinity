package mesh

import (
	"fmt"
	"github.com/squidboylan/go-affinity/functional/utils"
	"github.com/project-receptor/receptor/pkg/backends"
	"github.com/project-receptor/receptor/pkg/netceptor"
	"io/ioutil"
	"net"
	"os"
	_ "sync"
	"time"
	"gopkg.in/yaml.v2"
	"strconv"
)

type Node struct {
	netceptorInstance *netceptor.Netceptor
}

type Mesh struct {
	nodes map[string]Node
	connections map[string]string
}

type YamlData struct {
	Nodes map[string] *struct {
		Connections []string
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
		netceptorInstance: n1,
	}
}

func (n *Node) TCPListen(address string) error {
	b1, err := backends.NewTCPListener(address, nil)
	if err != nil {
		return err
	}
	n.netceptorInstance.RunBackend(b1, 1.0, handleError)
	return err
}

func (n *Node) TCPDial(address string) error {
	b1, err := backends.NewTCPDialer(address, true, nil)
	if err != nil {
		return err
	}
	n.netceptorInstance.RunBackend(b1, 1.0, handleError)
	return err
}

func (n *Node) ServiceListen(name string, function func(*netceptor.Listener)) (*netceptor.Listener, error) {
	l1, err := n.netceptorInstance.Listen(name, nil)
	if err != nil {
		return nil, err
	}
	go function(l1)
	return l1, err
}

func (n *Node) ServiceDial(node string, servicename string, timeout int, function func()) (net.Conn, error) {
	for timeout > 0 {
		fmt.Printf("Dialing node1\n")
		c2, err := n.netceptorInstance.Dial("node1", "echo", nil)
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
	nodes := make(map[string]Node)
	connections := make(map[string]string)

	yamlDat, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("failed to read %s", filename)
		os.Exit(1)
	}

	data := YamlData {}

	yaml.Unmarshal(yamlDat, &data)

	for k := range data.Nodes {
		node := NewNode(data.Nodes[k].Name)
		if data.Nodes[k].Listen != "" {
			fmt.Println(data.Nodes[k].Listen)
			err := node.TCPListen(data.Nodes[k].Listen)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			
		} else {
			retries := 5
			for retries > 0 {
				port := utils.RandomPort()
				addrString := "127.0.0.1:" + strconv.Itoa(port)
				fmt.Println(addrString)
				err := node.TCPListen(addrString)
				if err == nil {
					data.Nodes[k].Listen = addrString
					break
				}
				retries -= 1
			}
			if retries == 0 {
				fmt.Println("Failed to conenct to a port after trying 5 times")
				os.Exit(1)
			}
		}
		nodes[data.Nodes[k].Name] = node
	}
	for k := range data.Nodes {
		node := nodes[data.Nodes[k].Name]
		for _, conn := range data.Nodes[k].Connections {
			fmt.Println("data.Nodes[k].Name: " + data.Nodes[k].Name + " data.Nodes[k].Listen: " + data.Nodes[k].Listen)

			node.TCPDial(data.Nodes[conn].Listen)
			connections[data.Nodes[k].Name] = conn
		}
	}
	return Mesh {
		nodes,
		connections,
	}
}

// This is broken and causes the thread to hang, dont use until
// netceptor.Shutdown is fixed
func (m *Mesh)Shutdown() {
	for _, node := range m.nodes {
		node.netceptorInstance.Shutdown()
	}
}
