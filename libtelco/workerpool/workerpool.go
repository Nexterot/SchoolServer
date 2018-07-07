// Copyright (C) 2018 Mikhail Masyagin

/*
Package workerpool содержит "ООП-обертку" над пулом "воркеров" -
набором go-routine, постоянно находящихся в ожидании новых задач.
Использование пула связано с необходимостью экономить ресуры
ОЗУ и ЦП, а следовательно, переиспользовать уже запущенные go-routine'ы.
*/
package workerpool

import "sync"

// Task interface - описание интрефейса задачи.
type Task interface {
	Execute()
}

// WorkerPool struct содержит основную информацию о пуле "воркеров":
// - размер пула;
// - канал с набором тасков;
// - синхронизатор в виде Wait-группы;
type WorkerPool struct {
	size  int
	tasks chan Task
	wg    sync.WaitGroup
}

// NewWorkerPool создает новый пул "воркеров".
func NewWorkerPool(size int) *WorkerPool {
	pool := &WorkerPool{
		size:  size,
		tasks: make(chan Task, 128),
	}
	for n := 0; n < size; n++ {
		pool.wg.Add(1)
		go pool.NewWorker()
	}

	return pool
}

// NewWorker создает новый "воркер".
func (pool *WorkerPool) NewWorker() {
	defer pool.wg.Done()
	/*
		for {
			select {
			// Если есть задача, то ее нужно обработать
			case task, ok := <-pool.Tasks:
				if !ok {
					return
				}
				task.Execute()
			}
		}
	*/
	for task := range pool.tasks {
		task.Execute()
	}
}

// Close закрывает канал задач пула "воркеров".
func (pool *WorkerPool) Close() {
	close(pool.tasks)
}

// Wait ожидает завершения работы всех "воркеров".
func (pool *WorkerPool) Wait() {
	pool.wg.Wait()
}

// Execute запускает очередную задачу на выполнение.
func (pool *WorkerPool) Execute(task Task) {
	pool.tasks <- task
}
