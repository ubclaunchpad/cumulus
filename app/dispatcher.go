package app

// WorkerQueue is the channel of workers handling work.
var WorkerQueue chan GenericWorker

// QuitChan is the channel we use to kill the workers.
var QuitChan chan bool
