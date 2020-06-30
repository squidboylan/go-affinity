package mesh

import (
	_ "github.com/project-receptor/receptor/pkg/backends"
	_ "github.com/project-receptor/receptor/pkg/netceptor"
	"io/ioutil"
	"fmt"
	"os"
	_ "sync"
	"gopkg.in/yaml.v2"
	"time"
	"testing"
	"reflect"
)

func TestNode(t *testing.T) {
	testTable := []struct {
		filename string
	} {
		{"tree-mesh.yaml"},
		{"random-mesh.yaml"},
		{"flat-mesh.yaml"},
	}
	t.Parallel()
	for _, data := range testTable {
		filename := data.filename
		t.Run(filename, func(t *testing.T) {
			t.Parallel()
			mesh := NewMeshFromFile(filename)
			// We need to sleep for a bit so everything can advertise routes
			// and the routing table can settle
			time.Sleep(5 * time.Second)
			for _, status := range mesh.Status() {
				t.Log(status.NodeID)
				t.Log(status.RoutingTable)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	testTable := []struct {
		filename string
	} {
		{"tree-mesh.yaml"},
		{"random-mesh.yaml"},
		{"flat-mesh.yaml"},
	}
	t.Parallel()
	for _, data := range testTable {
		filename := data.filename
		t.Run(filename, func(t *testing.T) {
			t.Parallel()
			mesh := NewMeshFromFile(filename)
			yamlDat, err := ioutil.ReadFile(filename)
			if err != nil {
				fmt.Printf("failed to read %s", filename)
				os.Exit(1)
			}

			data := YamlData {}

			yaml.Unmarshal(yamlDat, &data)
			// We need to sleep for a bit so connections can happen
			time.Sleep(5 * time.Second)
			for _, status := range mesh.Status() {
				actualConnections := map[string]float64{}
				for _, connection := range status.Connections {
					actualConnections[connection.NodeID] = connection.Cost
				}
				expectedConnections := map[string]float64{}
				for k, v := range data.Nodes[status.NodeID].Connections {
					expectedConnections[k] = v
				}
				for nodeID, node := range data.Nodes {
					if nodeID == status.NodeID {
						continue
					}
					for k, v := range node.Connections {
						if k == status.NodeID {
							expectedConnections[nodeID] = v
						}
					}
				}
				if !reflect.DeepEqual(actualConnections, expectedConnections) {
					t.Error("Expected connections did not match actual connections for node: " + status.NodeID)
				}
			}
		})
	}
}
