package core

import (
	"log"
	"os"
	"sync"
)

type job struct {
	name  string
	dag   *Dag
	tasks map[string]*task
}

func Job(name string) *job {
	d := NewDag()
	j := job{name, d, make(map[string]*task)}
	return &j
}

func (j *job) addTask(t *task) {
	j.tasks[t.name] = t
	j.dag.addNode(t.name)
}

func (j *job) setDownstream(ind, dep *task) {
	j.dag.setDownstream(ind.name, dep.name)
}

func (j *job) run() error {
	if valid := j.dag.validate(); valid != true {
		return &InvalidDagError{}
	} else {
		err := j.run_tasks()
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (j *job) run_tasks() error {
	var wg sync.WaitGroup

	total := len(j.tasks)
	done := 0
	ind := j.dag.independentNodes()

	// Run the independent tasks
	for _, name := range ind {
		wg.Add(1)
		done += 1
		go j.tasks[name].run(&wg)
	}

	wg.Wait()

	// Run downstream tasks
	for {
		if done == total {
			break
		} else {
			// for each task
			for _, t := range j.tasks {
				if !t.isDone() {
					upstream_done := true
					// iterate over the dependencies
					for _, us := range j.dag.dependencies(t.name) {
						// if any upstream task is not done, set the flag to false
						if !j.tasks[us].isDone() {
							upstream_done = false
						}
					}

					if upstream_done {
						wg.Add(1)
						done += 1
						go t.run(&wg)
					}
				}
			}
		}

		wg.Wait()
	}

	return nil
}

type task struct {
	name     string
	status   string
	operator operator
}

func Task(name string, op operator) *task {
	t := task{name, "None", op}
	return &t
}

func (t *task) isDone() bool {
	if t.status == "Success" || t.status == "Failed" {
		return true
	} else {
		return false
	}
}

func (t *task) run(wg *sync.WaitGroup) error {
	defer wg.Done()
	var (
		logger = log.New(os.Stdout, "taskLogger:", log.Lshortfile)
	)

	res, err := t.operator.run()

	if err != nil {
		logger.Println("Task failed")
		t.status = "Failed"
		return err
	} else {
		logger.Println("Task", t.name, "succeeded with result", res)
		t.status = "Success"
		return nil
	}
}
