package mesh

import (
	_ "github.com/project-receptor/receptor/pkg/backends"
	_ "github.com/project-receptor/receptor/pkg/netceptor"
	_ "io"
	_ "net"
	_ "os"
	_ "sync"
	"time"
	"testing"
)

func TestNode(t *testing.T) {
	testTable := []struct {
		filename string
	} {
		{"tree-mesh.yaml"},
		{"random-mesh.yaml"},
		{"flat-mesh.yaml"},
	}
	for _, data := range testTable {
		filename := data.filename
		t.Run(filename, func(t *testing.T) {
			t.Parallel()
			mesh := NewMeshFromFile(filename)
			time.Sleep(5 * time.Second)
			for _, status := range mesh.Status() {
				t.Log(status.NodeID)
				t.Log(status.RoutingTable)
			}
		})
	}
}
