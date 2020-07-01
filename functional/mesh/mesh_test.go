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
	_ "reflect"
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
			//time.Sleep(1 * time.Second)
			mesh.WaitForReady(1000)
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
			timeout := 1000
			connectionsReady := false
			for ;timeout > 0 && !connectionsReady; connectionsReady = mesh.CheckConnections() {
				time.Sleep(200 * time.Millisecond)
				timeout -= 200
			}
			if connectionsReady == false {
				t.Error("Timed out while waiting for connections")
			}
		})
	}
}
