package executer

import (
	protocols2 "getitle/v1/pkg/nuclei/protocols"
)

type Executer struct {
	requests []protocols2.Request
	options  *protocols2.ExecuterOptions
}

type Event map[string]interface{}
type WrappedEvent struct {
	InternalEvent   Event
	OperatorsResult *protocols2.Result
}

var _ protocols2.Executer = &Executer{}

// NewExecuter creates a new request executer for list of requests
func NewExecuter(requests []protocols2.Request, options *protocols2.ExecuterOptions) *Executer {
	return &Executer{requests: requests, options: options}
}

// Compile compiles the execution generators preparing any requests possible.
func (e *Executer) Compile() error {
	for _, request := range e.requests {
		err := request.Compile(e.options)
		if err != nil {
			return err
		}
	}
	return nil
}

// Requests returns the total number of requests the rule will perform
func (e *Executer) Requests() int {
	var count int
	for _, request := range e.requests {
		count += request.Requests()
	}
	return count
}

// Execute executes the protocol group and returns true or false if results were found.
func (e *Executer) Execute(input string) (*protocols2.Result, error) {
	var result *protocols2.Result

	dynamicValues := make(map[string]interface{})
	//previous := make(map[string]interface{})
	for _, req := range e.requests {
		err := req.ExecuteWithResults(input, dynamicValues, func(event *protocols2.InternalWrappedEvent) {
			//ID := req.GetID()
			if event.OperatorsResult != nil {
				result = event.OperatorsResult
			}
		})
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
