package main

import (
	"math"
)

func costFunction(b float64, q1 float64, q2 float64) float64 {
	// reference: http://blog.oddhead.com/2006/10/30/implementing-hansons-market-maker/
	return b * math.Log(math.Pow(math.E, q1/b)+math.Pow(math.E, q2/b))
}

// logarithmic market scoring rule (LMSR) market maker from Robin Hanson:
// https://mason.gmu.edu/~rhanson/mktscore.pdf1
func BinaryLMSR(invariant int, funding int, q1 int, q2 int, dq1 int) float64 {
	b := float64(funding)
	fq1 := float64(q1)
	fq2 := float64(q2)
	fdq1 := float64(dq1)
	return costFunction(b, fq1+fdq1, fq2) - costFunction(b, fq1, fq2)
}
