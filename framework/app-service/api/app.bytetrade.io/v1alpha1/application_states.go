package v1alpha1

// ApplicationState is the state of an application at current time
type ApplicationState string

// These ar the valid states of applications
const (
	// AppRunning means that the application is installed success and ready for serve.
	AppRunning ApplicationState = "running"
	// AppStopped means that the application's deployment/statefulset replicas has been set to zero.
	AppStopped ApplicationState = "stopped"
	// AppNotReady means that the application's not ready to serve
	AppNotReady ApplicationState = "notReady"
)

func (a ApplicationState) String() string {
	return string(a)
}

/* ApplicationState change
+---------+  install   +-------------+     +------------+     +--------------+            +--------------+  suspend   +---------+  resume   +----------+
| pending | ---------> | downloading | --> | installing | --> | initializing | ---------> |              | ---------> | suspend | --------> | resuming |
+---------+            +-------------+     +------------+     +--------------+            |              |            +---------+           +----------+
                                                                                          |              |                                    |
                                                                +-----------------------> |   running    | <----------------------------------+
                                                                |                         |              |
                                                              +--------------+  upgrade   |              |
                                                              |  upgrading   | <--------- |              |
                                                              +--------------+            +--------------+
                                                                                            |
                                                                                            | install
                                                                                            v
                                                                                          +--------------+
                                                                                          | uninstalling |
                                                                                          +--------------+
*/
