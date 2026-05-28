// cafe-sim simulates a coffee shop using an event-driven approach.
// Events (arrivals, orders, completions) are processed in time order
// using a min-heap priority queue from github.com/lkorsman/pqueue.
//
// Usage:
//
// cafe-sim --customers 20 --baristas 2 --duration 60 --seed 42
package main

import (
	"flag"
	"fmt"
	"math/rand"

	"github.com/lkorsman/pqueue"
)

// eventType describes what is happening at a point in time.
type eventType int

const (
	eventArrive eventType = iota
	eventOrder
	eventFinish
)

// event represents something that happens at a specific time.
type event struct {
	time float64
	kind eventType
	customer string
	arriveAt float64 
}

type barista struct {
	name string
	freeAt float64
}

func main() {
	customers := flag.Int("customers", 20, "number of customers to simulate")
	baristas  := flag.Int("baristas", 2, "number of baristas working")
	duration  := flag.Float64("duration", 60, "how long the shop is open (minutes)")
	seed      := flag.Int64("seed", 42, "random seed for reproducibility")
	flag.Parse()
 
	run(*customers, *baristas, *duration, *seed)
}

func run(numCustomers int, numBaristas int, duration float64, seed int64) {
	rng := rand.New(rand.NewSource(seed))

	// min-heap ordered by event time - earliest event is always popped first
	queue := pqueue.New[event](func(a, b event) bool {
		return a.time < b.time
	})

	// schedule all customer arrivals upfront, spread across the opening hours
	for i := 1; i <= numCustomers; i++ {
		arrivalTime := rng.Float64() * duration
		queue.Push(event{
			time: arrivalTime,
			kind: eventArrive,
			customer: fmt.Sprintf("Customerv%d", i),
			arriveAt: arrivalTime,
		})
	}

	// set up baristas
	staff := make([]barista, numBaristas)
	for i := range staff {
		staff[i] = barista{
			name: fmt.Sprintf("Barista %d", i+1),
			freeAt: 0,
		}
	}

	// waiting customers (arrived but not yet being served)
	type waitingCustomer struct {
		name string
		arriveAt float64
	}
	waiting := []waitingCustomer{}
	
	fmt.Println("Cafe simulation starting")
	fmt.Printf("   %d customers, %d baristas, %.0f minute shift\n\n", numCustomers, numBaristas, duration)
	// process events in time order
	for !queue.IsEmpty() {
		e, _ := queue.Pop()
 
		switch e.kind {
		case eventArrive:
			fmt.Printf("[%s] %s arrives\n", formatTime(e.time), e.customer)
 
			// find a free barista
			b := firstFreeBarista(staff, e.time)
			if b != nil {
				// barista is free — start immediately
				orderTime := e.time
				brewTime  := 2 + rng.Float64()*4 // 2–6 minutes to make a drink
				b.freeAt   = orderTime + brewTime
 
				queue.Push(event{
					time:     orderTime,
					kind:     eventOrder,
					customer: e.customer,
					arriveAt: e.arriveAt,
				})
				queue.Push(event{
					time:     orderTime + brewTime,
					kind:     eventFinish,
					customer: e.customer,
					arriveAt: e.arriveAt,
				})
			} else {
				// all baristas busy — join the wait
				fmt.Printf("[%s] %s joins the queue\n", formatTime(e.time), e.customer)
				waiting = append(waiting, waitingCustomer{e.customer, e.arriveAt})
			}
 
		case eventOrder:
			// find which barista is serving this customer
			b := nextAvailableBarista(staff, e.time)
			if b != nil {
				fmt.Printf("[%s] %s starts making %s's order\n", formatTime(e.time), b.name, e.customer)
			}
 
		case eventFinish:
			wait := e.time - e.arriveAt
			fmt.Printf("[%s] %s's order is ready (waited %.1fm)\n", formatTime(e.time), e.customer, wait)
 
			// if someone is waiting, serve them now
			if len(waiting) > 0 {
				next    := waiting[0]
				waiting  = waiting[1:]
 
				b := firstFreeBarista(staff, e.time)
				if b == nil {
					// pick the barista who just finished
					b = justFinishedBarista(staff, e.time)
				}
 
				brewTime := 2 + rng.Float64()*4
				if b != nil {
					b.freeAt = e.time + brewTime
				}
 
				queue.Push(event{
					time:     e.time,
					kind:     eventOrder,
					customer: next.name,
					arriveAt: next.arriveAt,
				})
				queue.Push(event{
					time:     e.time + brewTime,
					kind:     eventFinish,
					customer: next.name,
					arriveAt: next.arriveAt,
				})
			}
		}
	}
 
	if len(waiting) > 0 {
		fmt.Printf("\n  %d customer(s) left without being served (shop closed)\n", len(waiting))
	}
 
	fmt.Println("\nSimulation complete")
}

// formatTime converts minutes-since-open into a readable clock time (starting at 08:00).
func formatTime(minutes float64) string {
	totalMinutes := int(minutes) + 8*60
	h := totalMinutes / 60
	m := totalMinutes % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}
 
// firstFreeBarista returns the first barista who is free at or before `at`.
func firstFreeBarista(staff []barista, at float64) *barista {
	for i := range staff {
		if staff[i].freeAt <= at {
			return &staff[i]
		}
	}
	return nil
}
 
// nextAvailableBarista returns the barista who became free most recently.
func nextAvailableBarista(staff []barista, at float64) *barista {
	return firstFreeBarista(staff, at)
}
 
// justFinishedBarista returns the barista whose freeAt is closest to `at`.
func justFinishedBarista(staff []barista, at float64) *barista {
	var closest *barista
	for i := range staff {
		if closest == nil || absDiff(staff[i].freeAt, at) < absDiff(closest.freeAt, at) {
			closest = &staff[i]
		}
	}
	return closest
}
 
func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}
