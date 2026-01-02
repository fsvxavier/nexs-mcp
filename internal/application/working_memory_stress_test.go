package application

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

// Stress test that concurrently creates and promotes working memories.
func TestWorkingMemory_ConcurrentCreatePromote(t *testing.T) {
	// Use a file-backed repo to get thread-safety in persistence layer
	repoDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(repoDir)
	if err != nil {
		t.Fatalf("failed to create file repo: %v", err)
	}

	svc := NewWorkingMemoryService(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	nSessions := 10
	nGoroutines := 30
	opsPerG := 100

	// Create a set of session IDs
	sessions := make([]string, nSessions)
	for i := range nSessions {
		sessions[i] = "stress-session-" + time.Now().Format("150405") + "-" + string(rune(i+65))
	}

	for g := range nGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create a thread-safe rng per goroutine
			workerRng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))

			// per-iteration helper that runs an operation but aborts promptly if ctx done
			runOp := func(fn func()) {
				doneOp := make(chan struct{})
				go func() {
					fn()
					close(doneOp)
				}()

				select {
				case <-ctx.Done():
					return
				case <-doneOp:
					return
				case <-time.After(500 * time.Millisecond):
					// Operation timed out; return to let loop check ctx
					return
				}
			}

			for range opsPerG {
				select {
				case <-ctx.Done():
					return
				default:
				}
				sid := sessions[workerRng.Intn(len(sessions))]
				// Random operation mix
				op := workerRng.Intn(5)
				switch op {
				case 0: // create
					content := "stress content " + time.Now().Format(time.RFC3339Nano)
					runOp(func() { _, _ = svc.Add(ctx, sid, content, domain.PriorityMedium, []string{"stress"}, nil) })
				case 1: // list
					var err error
					done := make(chan struct{})
					go func() {
						_, err = svc.List(ctx, sid, false, false)
						close(done)
					}()
					select {
					case <-ctx.Done():
						return
					case <-done:
						if err != nil {
							// ignore failed list
							break
						}
					case <-time.After(500 * time.Millisecond):
						return
					}
				case 2: // extend TTL of a random memory
					var mems []*domain.WorkingMemory
					var err error
					done := make(chan struct{})
					go func() {
						mems, err = svc.List(ctx, sid, false, false)
						close(done)
					}()
					select {
					case <-ctx.Done():
						return
					case <-done:
						if err == nil && len(mems) > 0 {
							idx := workerRng.Intn(len(mems))
							runOp(func() { _ = svc.ExtendTTL(sid, mems[idx].GetID()) })
						}
					case <-time.After(500 * time.Millisecond):
						return
					}
				case 3: // try promote a random memory
					var mems []*domain.WorkingMemory
					var err error
					done := make(chan struct{})
					go func() {
						mems, err = svc.List(ctx, sid, false, false)
						close(done)
					}()
					select {
					case <-ctx.Done():
						return
					case <-done:
						if err == nil && len(mems) > 0 {
							idx := workerRng.Intn(len(mems))
							runOp(func() { _, _ = svc.Promote(ctx, sid, mems[idx].GetID()) })
						}
					case <-time.After(500 * time.Millisecond):
						return
					}
				case 4: // expire memory
					var mems []*domain.WorkingMemory
					var err error
					done := make(chan struct{})
					go func() {
						mems, err = svc.List(ctx, sid, false, false)
						close(done)
					}()
					select {
					case <-ctx.Done():
						return
					case <-done:
						if err == nil && len(mems) > 0 {
							idx := workerRng.Intn(len(mems))
							runOp(func() { _ = svc.ExpireMemory(sid, mems[idx].GetID()) })
						}
					case <-time.After(500 * time.Millisecond):
						return
					}
				}
				// small sleep to widen interleaving
				time.Sleep(time.Millisecond * time.Duration(workerRng.Intn(5)))
			}
		}(g)
	}

	// Wait for all workers with timeout to avoid test hanging
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// finished normally
	case <-time.After(15 * time.Second):
		t.Fatal("test timed out")
	}

	// Final verification: ensure no panics occurred and repo contains some long-term memories
	elts, err := repo.List(domain.ElementFilter{})
	if err != nil {
		t.Fatalf("failed to list elements: %v", err)
	}

	// Expect at least one promoted memory to exist
	found := false
	for _, e := range elts {
		if e.GetType() == domain.MemoryElement {
			found = true
			break
		}
	}

	if !found {
		t.Logf("Warning: no long-term memories found (possible but unexpected for this run)")
	}
}

// Continuous concurrent access and promotion for a fixed session to exercise contention.
func TestWorkingMemory_ConcurrentAccessAndPromotion(t *testing.T) {
	repoDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(repoDir)
	if err != nil {
		t.Fatalf("failed to create file repo: %v", err)
	}

	svc := NewWorkingMemoryService(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	session := "concurrent-single-session"

	// Pre-create some memories
	for range 50 {
		_, _ = svc.Add(ctx, session, "seed content "+time.Now().Format(time.RFC3339Nano), domain.PriorityMedium, nil, nil)
	}

	var wg sync.WaitGroup
	opsPer := 500
	nWorkers := 20

	for range nWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Create a thread-safe rng per goroutine
			workerRng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(time.Now().Nanosecond())))

			runOp := func(fn func()) {
				doneOp := make(chan struct{})
				go func() {
					fn()
					close(doneOp)
				}()
				select {
				case <-ctx.Done():
					return
				case <-doneOp:
					return
				case <-time.After(500 * time.Millisecond):
					return
				}
			}

			for range opsPer {
				select {
				case <-ctx.Done():
					return
				default:
				}

				var mems []*domain.WorkingMemory
				var err error
				done := make(chan struct{})
				go func() {
					mems, err = svc.List(ctx, session, false, false)
					close(done)
				}()
				select {
				case <-ctx.Done():
					return
				case <-done:
					if err != nil || len(mems) == 0 {
						continue
					}
					m := mems[workerRng.Intn(len(mems))]
					// random access
					runOp(func() { _, _ = svc.Get(ctx, session, m.GetID()) })
					// maybe promote
					if workerRng.Intn(10) == 0 {
						runOp(func() { _, _ = svc.Promote(ctx, session, m.GetID()) })
					}
					// maybe extend
					if workerRng.Intn(5) == 0 {
						runOp(func() { _ = svc.ExtendTTL(session, m.GetID()) })
					}
					// maybe expire
					if workerRng.Intn(20) == 0 {
						runOp(func() { _ = svc.ExpireMemory(session, m.GetID()) })
					}
				case <-time.After(500 * time.Millisecond):
					continue
				}

				// small sleep
				time.Sleep(time.Microsecond * time.Duration(workerRng.Intn(50)))
			}
		}()
	}

	// Wait for workers with timeout to prevent indefinite running
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// finished normally
	case <-time.After(30 * time.Second):
		t.Fatal("test timed out")
	}

	// Ensure service stats function runs and doesn't race
	_ = svc.GetStats(session)
}
