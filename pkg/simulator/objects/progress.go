package objects

import (
	"fmt"
	"time"
)

type ProgressManager struct {
	name  string
	start time.Time
}

func NewProgressManager(name string) *ProgressManager {
	return &ProgressManager{
		name: name,
	}
}

func (p *ProgressManager) progress(current, total int) {
	fmt.Printf("[%s] %d/%d\r", p.name, current, total)

	if current == 1 {
		p.start = time.Now()
	}

	if current == total {
		fmt.Printf("\n")
		fmt.Printf("Time to load all objects: %s seconds\n\n", time.Since(p.start))
	}
}
