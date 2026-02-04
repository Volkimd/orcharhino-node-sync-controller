package controller

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const syncLabel = "orcharhino.de/synced"

// Datenstrukturen für die orcharhino API
type HostData struct {
	Name           string `json:"name"`
	HostgroupID    int    `json:"hostgroup_id,omitempty"`
	LocationID     int    `json:"location_id,omitempty"`
	OrganizationID int    `json:"organization_id,omitempty"`
	Build          bool   `json:"build"`
	Managed        bool   `json:"managed"`
}

type HostPayload struct {
	Host HostData `json:"host"`
}

type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Hilfsfunktion für den authentifizierten HTTP-Client
func (r *NodeReconciler) getAuthenticatedClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}
}

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=nodes/status,verbs=get

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	node := &corev1.Node{}

	err := r.Get(ctx, req.NamespacedName, node)

	if apierrors.IsNotFound(err) {
		l.Info("Node gelöscht, starte orcharhino Cleanup", "node", req.Name)
		return ctrl.Result{}, r.deleteFromOrcharhino(req.Name)
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// Nur syncen, wenn Label noch nicht auf true
	if node.Labels[syncLabel] != "true" {
		l.Info("Neuer Node erkannt, registriere in orcharhino", "node", node.Name)

		if err := r.syncToOrcharhino(node); err != nil {
			l.Error(err, "Fehler beim orcharhino Sync")
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}

		if node.Labels == nil {
			node.Labels = make(map[string]string)
		}
		node.Labels[syncLabel] = "true"
		if err := r.Update(ctx, node); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) syncToOrcharhino(node *corev1.Node) error {
	url := os.Getenv("ORCHARHINO_URL")
	user := os.Getenv("ORCHARHINO_USER")
	pass := os.Getenv("ORCHARHINO_PASS")

	payload := HostPayload{
		Host: HostData{
			Name:    node.Name,
			Build:   false,
			Managed: false,
			// Hier ggf. IDs ergänzen, z.B. HostgroupID: 1
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	apiPath := fmt.Sprintf("%s/api/hosts", url)
	req, err := http.NewRequest("POST", apiPath, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")

	client := r.getAuthenticatedClient()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("orcharhino API Error: Status %d", resp.StatusCode)
	}
	return nil
}

func (r *NodeReconciler) deleteFromOrcharhino(nodeName string) error {
	url := os.Getenv("ORCHARHINO_URL")
	user := os.Getenv("ORCHARHINO_USER")
	pass := os.Getenv("ORCHARHINO_PASS")

	apiPath := fmt.Sprintf("%s/api/hosts/%s", url, nodeName)
	req, err := http.NewRequest("DELETE", apiPath, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(user, pass)

	client := r.getAuthenticatedClient()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("orcharhino Delete Error: Status %d", resp.StatusCode)
	}
	return nil
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
