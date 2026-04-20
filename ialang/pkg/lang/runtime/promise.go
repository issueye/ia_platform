package runtime

import "context"

type Promise struct {
	done  chan struct{}
	value Value
	err   error
}

var _ Awaitable = (*Promise)(nil)

func NewPromise(task AsyncTask) *Promise {
	p := &Promise{done: make(chan struct{})}
	go func() {
		p.value, p.err = task()
		close(p.done)
	}()
	return p
}

func ResolvedPromise(v Value) *Promise {
	p := &Promise{
		done:  make(chan struct{}),
		value: v,
	}
	close(p.done)
	return p
}

func (p *Promise) Await() (Value, error) {
	<-p.done
	return p.value, p.err
}

func (p *Promise) AwaitContext(ctx context.Context) (Value, error) {
	select {
	case <-p.done:
		return p.value, p.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Promise) IsDone() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}

// PromiseAll creates a promise that resolves when all input promises resolve.
// Returns an array of results in the same order as input.
// If any promise rejects, the returned promise rejects immediately.
func PromiseAll(promises []Awaitable) Awaitable {
	return NewPromise(func() (Value, error) {
		results := make(Array, len(promises))
		for i, p := range promises {
			v, err := p.Await()
			if err != nil {
				return nil, err
			}
			results[i] = v
		}
		return results, nil
	})
}

// PromiseRace creates a promise that resolves or rejects as soon as one of the
// input promises resolves or rejects, with the value or error from that promise.
func PromiseRace(promises []Awaitable) Awaitable {
	return NewPromise(func() (Value, error) {
		type result struct {
			value Value
			err   error
		}
		done := make(chan result, 1)
		for _, p := range promises {
			go func(awaitable Awaitable) {
				v, err := awaitable.Await()
				select {
				case done <- result{value: v, err: err}:
				default:
				}
			}(p)
		}
		r := <-done
		return r.value, r.err
	})
}

// PromiseAllSettled creates a promise that resolves when all input promises
// have settled (either resolved or rejected). Returns an array of objects
// with {status, value?, reason?}.
func PromiseAllSettled(promises []Awaitable) Awaitable {
	return NewPromise(func() (Value, error) {
		results := make(Array, len(promises))
		for i, p := range promises {
			v, err := p.Await()
			if err != nil {
				results[i] = Object{
					"status": "rejected",
					"reason": err.Error(),
				}
			} else {
				results[i] = Object{
					"status": "fulfilled",
					"value":  v,
				}
			}
		}
		return results, nil
	})
}
