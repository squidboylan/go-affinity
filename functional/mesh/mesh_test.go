package mesh

import (
	_ "github.com/project-receptor/receptor/pkg/backends"
	_ "github.com/project-receptor/receptor/pkg/netceptor"
	"io/ioutil"
	_ "sync"
	"gopkg.in/yaml.v2"
	"time"
	"testing"
	_ "reflect"
)

func TestNode(t *testing.T) {
	testTable := []struct {
		filename string
	} {
		{"mesh-definitions/flat-mesh-tcp.yaml"},
		{"mesh-definitions/random-mesh-tcp.yaml"},
		{"mesh-definitions/tree-mesh-tcp.yaml"},
		{"mesh-definitions/flat-mesh-udp.yaml"},
		{"mesh-definitions/random-mesh-udp.yaml"},
		{"mesh-definitions/tree-mesh-udp.yaml"},
	}
	t.Parallel()
	for _, data := range testTable {
		filename := data.filename
		t.Run(filename, func(t *testing.T) {
			t.Parallel()
			mesh, err := NewMeshFromFile(filename)
			if err != nil {
				t.Error(err)
			}
			err = mesh.WaitForReady(10000)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	testTable := []struct {
		filename string
	} {
		{"mesh-definitions/flat-mesh-tcp.yaml"},
		{"mesh-definitions/random-mesh-tcp.yaml"},
		{"mesh-definitions/tree-mesh-tcp.yaml"},
		{"mesh-definitions/flat-mesh-udp.yaml"},
		{"mesh-definitions/random-mesh-udp.yaml"},
		{"mesh-definitions/tree-mesh-udp.yaml"},
	}
	t.Parallel()
	for _, data := range testTable {
		filename := data.filename
		t.Run(filename, func(t *testing.T) {
			t.Parallel()
			mesh, err := NewMeshFromFile(filename)
			if err != nil {
				t.Error(err)
			}
			yamlDat, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Error(err)
			}

			data := YamlData {}

			yaml.Unmarshal(yamlDat, &data)
			// We need to sleep for a bit so connections can happen
			timeout := 1000
			connectionsReady := false
			for ;timeout > 0 && !connectionsReady; connectionsReady = mesh.CheckConnections() {
				time.Sleep(100 * time.Millisecond)
				timeout -= 100
			}
			if connectionsReady == false {
				t.Error("Timed out while waiting for connections")
			}
		})
	}
}

func BenchmarkLinearMeshStartup(b *testing.B) {
	// Setup our mesh yaml data
	totalNodes := 100
	data := YamlData {}
	data.Nodes = make(map[string]*YamlNode)

	for i := 0; i < totalNodes; i++ {
		connections := make(map[string]float64)
		nodeID := "Node" + string(i)
		if i > 0 {
			prevNodeID := "Node" + string(i-1)
			connections[prevNodeID] = 1
		}
		data.Nodes[nodeID] = &YamlNode {
			Connections: connections,
			Listen: []*YamlListener {
				&YamlListener {
					Addr: "",
					Cost: 1,
					Protocol: "tcp",
				},
			},
			Name: nodeID,
			Stats_enable: false,
			Stats_port: "",
		}
	}

	// Reset the Timer because building the yaml data for the mesh may have
	// taken a bit
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We probably dont need to stop the timer for this
		b.StopTimer()
		for k, _ := range data.Nodes {
			for _, listener := range data.Nodes[k].Listen {
				listener.Addr = ""
			}
		}
		b.StartTimer()
		mesh, err := NewMeshFromYaml(&data)
		if err != nil {
			b.Error(err)
		}
		err = mesh.WaitForReady(10000)
		if err != nil {
			b.Error(err)
		}
	}
}
