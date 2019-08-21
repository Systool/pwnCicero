package main

import (
	"fmt"
	"time"
)

type TransportSolution struct {
	ContextID    string
	SolNum       string
	TripDuration time.Duration
	TripLength   int
}

func (tpsol *TransportSolution) fromJson(jobj map[string]interface{}) error {
	if tpsol == nil {
		tpsol = &TransportSolution{}
	}

	tpsol.ContextID = jobj["IdContesto"].(string)
	tpsol.SolNum = string(jobj["Numero"].(int))
	tpsol.TripLength = jobj["MetriTotali"].(int)

	duration, err := time.ParseDuration(fmt.Sprint(jobj["MinutiTotali"].(int)) + "m")
	if err != nil {
		return err
	}
	tpsol.TripDuration = duration

	return nil
}
