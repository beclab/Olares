package middlewarerequest

import (
	"fmt"
	"io"
	"net/http"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	wrabbit "bytetrade.io/web3os/tapr/pkg/workload/rabbitmq"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

const rabbitMQNs = "rabbitmq-middleware"

func (c *controller) createOrUpdateRabbitMQRequest(req *aprv1.MiddlewareRequest) error {
	rmqc, err := c.newRabbitMQClient()
	if err != nil {
		klog.Errorf("failed to new rabbit client %v", err)
		return err
	}

	userPassword, err := req.Spec.RabbitMQ.Password.GetVarValue(c.ctx, c.k8sClientSet, req.Namespace)
	if err != nil {
		klog.Errorf("failed to get user password %v", err)
		return err
	}

	err = c.createOrUpdateRabbitUser(rmqc, req.Spec.RabbitMQ.User, userPassword)
	if err != nil {
		klog.Errorf("failed to create or update rabbitmq user %s, %v", req.Spec.RabbitMQ.User, err)
		return err
	}

	for _, v := range req.Spec.RabbitMQ.Vhosts {
		vhost := wrabbit.GetVhostName(req.Spec.AppNamespace, v.Name)
		err = c.ensureRabbitVhost(rmqc, vhost)
		if err != nil {
			klog.Errorf("failed to ensure rabbitmq vhost %s %v", vhost, err)
			return err
		}
		err = c.setRabbitPermissions(rmqc, req.Spec.RabbitMQ.User, vhost)
		if err != nil {
			klog.Errorf("failed to set rabbitmq vhost %s permission %v", vhost, err)
			return err
		}
	}
	return nil
}

func (c *controller) getRabbitMqEndpoint() string {
	return fmt.Sprintf("http://rabbitmq-rabbitmq.%s:15672", rabbitMQNs)
}

func (c *controller) newRabbitMQClient() (*rabbithole.Client, error) {
	adminUser, adminPassword, err := wrabbit.FindRabbitMQAdminUser(c.ctx, c.k8sClientSet, rabbitMQNs)
	if err != nil {
		klog.Errorf("failed to get root user info %v", err)
		return nil, err
	}

	endpoint := c.getRabbitMqEndpoint()
	rmqc, err := rabbithole.NewClient(endpoint, adminUser, adminPassword)
	if err != nil {
		klog.Errorf("failed to new rabbitmq client %v", err)
		return nil, err
	}
	return rmqc, nil
}

func (c *controller) deleteRabbitMQRequest(req *aprv1.MiddlewareRequest) error {
	rmqc, err := c.newRabbitMQClient()
	if err != nil {
		klog.Errorf("failed to new rabbit client %v", err)
		if apierrors.IsNotFound(err) {
			// RabbitMQ admin secret missing, service likely already removed. No-op.
			klog.Infof("rabbitmq admin secret not found, skipping deletion for user %s", req.Spec.RabbitMQ.User)
			return nil
		}
		return err
	}
	user := req.Spec.RabbitMQ.User
	for _, v := range req.Spec.RabbitMQ.Vhosts {
		vhost := wrabbit.GetVhostName(req.Spec.AppNamespace, v.Name)
		err := c.deleteRabbitPermissions(rmqc, user, vhost)
		if err != nil {
			klog.Errorf("failed to delete rabbit permissions user %s vhost %s %v", user, vhost, err)
			return err
		}
		err = c.deleteRabbitVhost(rmqc, vhost)
		if err != nil {
			return fmt.Errorf("failed to delete vhost %s %v", vhost, err)
		}
	}
	err = c.deleteRabbitUser(rmqc, user)
	if err != nil {
		return fmt.Errorf("failed to delete rabbitmq user %s %v", user, err)
	}
	return nil
}

func (c *controller) ensureRabbitVhost(client *rabbithole.Client, vhost string) error {
	resp, err := client.PutVhost(vhost, rabbithole.VhostSettings{})
	if err != nil {
		return fmt.Errorf("failed to put vhost %s, %v", vhost, err)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create vhost failed %s", string(bodyBytes))
	}
	return nil
}

func (c *controller) createOrUpdateRabbitUser(client *rabbithole.Client, username, password string) error {
	resp, err := client.PutUser(username, rabbithole.UserSettings{Password: password})
	if err != nil {
		klog.Errorf("failed to put user %s, %v", username, err)
		return err
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create user failed %s", string(bodyBytes))
	}
	return nil
}

func (c *controller) setRabbitPermissions(client *rabbithole.Client, username, vhost string) error {
	resp, err := client.UpdatePermissionsIn(vhost, username, rabbithole.Permissions{Configure: ".*", Write: ".*", Read: ".*"})
	if err != nil {
		return fmt.Errorf("failed to update vhost %s user %s %v", vhost, username, err)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set permission failed %s", string(bodyBytes))
	}
	return nil
}

func (c *controller) deleteRabbitPermissions(client *rabbithole.Client, username, vhost string) error {
	resp, err := client.ClearPermissionsIn(vhost, username)
	if err != nil {
		return fmt.Errorf("failed to clear permission user %s vhost %s %v", username, vhost, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete permission failed %s", string(bodyBytes))
	}
	return nil
}

func (c *controller) deleteRabbitUser(client *rabbithole.Client, username string) error {
	resp, err := client.DeleteUser(username)
	if err != nil {
		return fmt.Errorf("failed to delete user %s %v", username, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete user %s failed %s", username, string(bodyBytes))
	}
	return nil
}

func (c *controller) deleteRabbitVhost(client *rabbithole.Client, vhost string) error {
	resp, err := client.DeleteVhost(vhost)
	if err != nil {
		return fmt.Errorf("failed to delete vhost %s %v", vhost, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete vhost %s failed %s", vhost, string(bodyBytes))
	}
	return nil
}
