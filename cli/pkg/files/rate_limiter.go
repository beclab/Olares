package files

import (
	"context"
	"time"

	"github.com/cavaliergopher/grab/v3"
)

type RateLimiter struct {
	r, n int
}

func NewLimiter(r int) grab.RateLimiter {
	return &RateLimiter{r: r}
}

func (c *RateLimiter) WaitN(ctx context.Context, n int) (err error) {
	c.n += n
	time.Sleep(
		time.Duration(1.00 / float64(c.r) * float64(n) * float64(time.Second)))
	return
}
