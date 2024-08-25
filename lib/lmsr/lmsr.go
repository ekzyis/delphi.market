package lmsr

import "math"

// Robin Hanson's Logarithmic Market Scoring Rule (LMSR) market maker
// https://mason.gmu.edu/~rhanson/mktscore.pdf

// "The market maker keeps track of how many shares have been purchased by
// traders in total so far for each outcome: that is, the number of shares
// outstanding for each outcome. Let q1 and q2 be the number (“quantity”)
// of shares outstanding for each of the two outcomes. The market maker also
// maintains a cost function C(q1,q2) which records how much money traders
// have collectively spent so far, and depends only on the number of shares
// outstanding, q1 and q2. For LMSR, the cost function is [this]:
//
// <see code>
// ...
//
// The parameter “b” controls the maximum possible amount of money the market
// maker can lose (which happens to be b*ln2 in the two-outcome case). The
// larger “b” is, the more money the market maker can lose. But a larger “b”
// also means the market has more liquidity or depth, meaning that traders can
// buy more shares at or near the current price without causing massive price swings."
//
// -- David Pennock, http://blog.oddhead.com/2006/10/30/implementing-hansons-market-maker/
func binaryLMSRcost(b float64, q1 float64, q2 float64) float64 {
	return b * math.Log(math.Pow(math.E, q1/b)+math.Pow(math.E, q2/b))
}

func Quote(b float64, q1 int, q2 int, dq1 int) float64 {
	// q1 must always be the quantity of shares that are bought / sold.
	// If you want to change q2, you need to swap q1 and q2 as input arguments.
	fq1 := float64(q1)
	fq2 := float64(q2)
	fdq1 := float64(dq1)
	return binaryLMSRcost(b, fq1+fdq1, fq2) - binaryLMSRcost(b, fq1, fq2)
}

func Price(b float64, q1 int, q2 int) float64 {
	//
	// q1 must always be the quantity of shares that are bought / sold.
	// If you want to change q2, you need to swap q1 and q2 as input arguments.
	//
	// An infinitesimal share would cost this much:
	//
	//   p = e^{q1/b} / ( e^{q1/b} + e^{q2/b} )
	//
	// However, to find the price of buying one full share
	// we need to use the cost function.
	return Quote(b, q1, q2, 1)
}
