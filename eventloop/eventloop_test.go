package eventloop

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
)

func TestRun(t *testing.T) {
	t.Parallel()
	const SCRIPT = `
	setTimeout(function() {
		console.log("ok");
	}, 1000);
	console.log("Started");
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		vm.RunProgram(prg)
	})
}

func TestStart(t *testing.T) {
	t.Parallel()
	const SCRIPT = `
	setTimeout(function() {
		console.log("ok");
	}, 1000);
	console.log("Started");
	`

	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}

	loop := NewEventLoop()
	loop.Start()

	loop.RunOnLoop(func(vm *goja.Runtime) {
		vm.RunProgram(prg)
	})

	time.Sleep(2 * time.Second)
	loop.Stop()
}

func TestInterval(t *testing.T) {
	t.Parallel()
	const SCRIPT = `
	var count = 0;
	var t = setInterval(function() {
		console.log("tick");
		if (++count > 2) {
			clearInterval(t);
		}
	}, 1000);
	console.log("Started");
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		vm.RunProgram(prg)
	})
}

func TestRunNoSchedule(t *testing.T) {
	loop := NewEventLoop()
	fired := false
	loop.Run(func(vm *goja.Runtime) { // should not hang
		fired = true
		// do not schedule anything
	})

	if !fired {
		t.Fatal("Not fired")
	}
}

func TestRunWithConsole(t *testing.T) {
	const SCRIPT = `
	console.log("Started");
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	if err != nil {
		t.Fatal("Call to console.log generated an error", err)
	}

	loop = NewEventLoop(EnableConsole(true))
	prg, err = goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	if err != nil {
		t.Fatal("Call to console.log generated an error", err)
	}
}

func TestRunNoConsole(t *testing.T) {
	const SCRIPT = `
	console.log("Started");
	`

	loop := NewEventLoop(EnableConsole(false))
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	if err == nil {
		t.Fatal("Call to console.log did not generate an error", err)
	}
}

func TestClearIntervalRace(t *testing.T) {
	t.Parallel()
	const SCRIPT = `
	console.log("calling setInterval");
	var t = setInterval(function() {
		console.log("tick");
	}, 500);
	console.log("calling sleep");
	sleep(2000);
	console.log("calling clearInterval");
	clearInterval(t);
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	// Should not hang
	loop.Run(func(vm *goja.Runtime) {
		vm.Set("sleep", func(ms int) {
			<-time.After(time.Duration(ms) * time.Millisecond)
		})
		vm.RunProgram(prg)
	})
}

func TestNativeTimeout(t *testing.T) {
	t.Parallel()
	fired := false
	loop := NewEventLoop()
	loop.SetTimeout(func(*goja.Runtime) {
		fired = true
	}, 1*time.Second)
	loop.Run(func(*goja.Runtime) {
		// do not schedule anything
	})
	if !fired {
		t.Fatal("Not fired")
	}
}

func TestNativeClearTimeout(t *testing.T) {
	t.Parallel()
	fired := false
	loop := NewEventLoop()
	timer := loop.SetTimeout(func(*goja.Runtime) {
		fired = true
	}, 2*time.Second)
	loop.SetTimeout(func(*goja.Runtime) {
		loop.ClearTimeout(timer)
	}, 1*time.Second)
	loop.Run(func(*goja.Runtime) {
		// do not schedule anything
	})
	if fired {
		t.Fatal("Cancelled timer fired!")
	}
}

func TestNativeInterval(t *testing.T) {
	t.Parallel()
	count := 0
	loop := NewEventLoop()
	var i *Interval
	i = loop.SetInterval(func(*goja.Runtime) {
		t.Log("tick")
		count++
		if count > 2 {
			loop.ClearInterval(i)
		}
	}, 1*time.Second)
	loop.Run(func(*goja.Runtime) {
		// do not schedule anything
	})
	if count != 3 {
		t.Fatal("Expected interval to fire 3 times, got", count)
	}
}

func TestNativeClearInterval(t *testing.T) {
	t.Parallel()
	count := 0
	loop := NewEventLoop()
	loop.Run(func(*goja.Runtime) {
		i := loop.SetInterval(func(*goja.Runtime) {
			t.Log("tick")
			count++
		}, 500*time.Millisecond)
		<-time.After(2 * time.Second)
		loop.ClearInterval(i)
	})
	if count != 0 {
		t.Fatal("Expected interval to fire 0 times, got", count)
	}
}

func TestSetTimeoutConcurrent(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop()
	loop.Start()
	ch := make(chan struct{}, 1)
	loop.SetTimeout(func(*goja.Runtime) {
		ch <- struct{}{}
	}, 100*time.Millisecond)
	<-ch
	loop.Stop()
}

func TestClearTimeoutConcurrent(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop()
	loop.Start()
	timer := loop.SetTimeout(func(*goja.Runtime) {
	}, 100*time.Millisecond)
	loop.ClearTimeout(timer)
	loop.Stop()
	if c := loop.jobCount; c != 0 {
		t.Fatalf("jobCount: %d", c)
	}
}

func TestClearIntervalConcurrent(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop()
	loop.Start()
	ch := make(chan struct{}, 1)
	i := loop.SetInterval(func(*goja.Runtime) {
		ch <- struct{}{}
	}, 500*time.Millisecond)

	<-ch
	loop.ClearInterval(i)
	loop.Stop()
	if c := loop.jobCount; c != 0 {
		t.Fatalf("jobCount: %d", c)
	}
}

func TestRunOnStoppedLoop(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop()
	var failed int32
	done := make(chan struct{})
	go func() {
		for atomic.LoadInt32(&failed) == 0 {
			loop.Start()
			time.Sleep(10 * time.Millisecond)
			loop.Stop()
		}
	}()
	go func() {
		for atomic.LoadInt32(&failed) == 0 {
			loop.RunOnLoop(func(*goja.Runtime) {
				if !loop.canRun {
					atomic.StoreInt32(&failed, 1)
					close(done)
					return
				}
			})
			time.Sleep(10 * time.Millisecond)
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	if atomic.LoadInt32(&failed) != 0 {
		t.Fatal("running job on stopped loop")
	}
}

func TestRunWithBuffer(t *testing.T) {
	const SCRIPT = `
	Buffer.from([1, 2, 3]);
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	if err != nil {
		t.Fatal("Call to Buffer.from generated an error", err)
	}

	loop = NewEventLoop()
	prg, err = goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	if err != nil {
		t.Fatal("Call to Buffer.From generated an error", err)
	}
}
