package calculate_curve_correlation

import "math"

// CalculateCurveCorrelation 计算两条曲线的相似度
// 返回值在区间： [-1,1]
// 如返回-10，则证明输入参数无效
// 原始出处不详，《数学之美：判定两个随机信号序列的相似度》
func CalculateCurveCorrelation(s1, s2 []float64, n int) float64 {
	var sum_s12 = 0.0
	var sum_s1 = 0.0
	var sum_s2 = 0.0
	var sum_s1s1 = 0.0 //s1^2
	var sum_s2s2 = 0.0 //s2^2
	var pxy = 0.0
	var temp1 = 0.0
	var temp2 = 0.0

	if s1 == nil || s2 == nil || n <= 0 {
		return -10
	}

	for i := 0; i < n; i++ {
		sum_s12 += s1[i] * s2[i]
		sum_s1 += s1[i]
		sum_s2 += s2[i]
		sum_s1s1 += s1[i] * s1[i]
		sum_s2s2 += s2[i] * s2[i]
	}

	temp1 = float64(n)*sum_s1s1 - sum_s1*sum_s1
	temp2 = float64(n)*sum_s2s2 - sum_s2*sum_s2

	if (temp1 > -delta && temp1 < delta) ||
		(temp2 > -delta && temp2 < delta) ||
		(temp1*temp2 <= 0) {
		return -10
	}

	pxy = (float64(n)*sum_s12 - sum_s1*sum_s2) / math.Sqrt(temp1*temp2)

	return pxy
}

const delta = 0.0001
