package apiserver

import (
	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers/custom"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"k8s.io/klog/v2"
)

type Server struct {
	eventWatchers *watchers.Watchers
	app           *fiber.App
	subscriber    *custom.Subscriber
}

func NewServer(w *watchers.Watchers, n *watchers.Notification) *Server {

	s := &Server{eventWatchers: w, subscriber: (&custom.Subscriber{}).WithNotification(n)}
	// create new fiber instance  and use across whole app
	app := fiber.New()

	// middleware to allow all clients to communicate using http and allow cors
	app.Use(cors.New())

	app.Post("/events/fire", s.fireEvent)

	s.app = app

	return s
}

func (s *Server) Run() {
	klog.Fatal(s.app.Listen(":8080"))
}

func (s *Server) ShutDown() {
	s.app.Shutdown()
}

func (s *Server) fireEvent(ctx *fiber.Ctx) error {
	event := &custom.CustomEvent{}
	err := ctx.BodyParser(event)
	if err != nil {
		klog.Error("read event body error, ", err)
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	s.eventWatchers.Enqueue(
		watchers.EnqueueObj{
			Obj:       event,
			Action:    watchers.ADD,
			Subscribe: s.subscriber,
		},
	)

	ctx.JSON(fiber.Map{
		"code": 0,
		"msg":  "success",
	})
	return nil
}
