package sub_timeline_fixer

import (
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"gonum.org/v1/gonum/cmplxs"
	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/floats"
	"math"
)

/*
	复现 https://github.com/smacke/ffsubsync 的 FFTAligner 算法
*/
type FFTAligner struct {
	maxOffsetSamples int
}

func NewFFTAligner(maxOffsetSeconds, sampleRate int) *FFTAligner {

	maxOffsetSamples := maxOffsetSeconds * sampleRate
	if maxOffsetSamples < 0 {
		maxOffsetSamples = -maxOffsetSamples
	}
	return &FFTAligner{
		maxOffsetSamples: maxOffsetSamples,
	}
}

// Fit 给出最佳的偏移，还需要根据实际情况进行转换（比如，1 步 是 10 ms）,输入的数组只能是 1 -1 这样的值，需要在外部做好归一化
func (f FFTAligner) Fit(refFloats, subFloats []float64) (int, float64) {

	convolve := f.fit(refFloats, subFloats)
	return f.computeArgmax(f.eliminateExtremeOffsetsFromSolutions(convolve, subFloats), subFloats)
}

// fit 返回 convolve
func (f FFTAligner) fit(refFloats, subFloats []float64) []float64 {

	// 计算出一维矩阵的长度
	totalBits := math.Log2(float64(len(refFloats)) + float64(len(subFloats)))
	totalLength := int(math.Pow(2, math.Ceil(totalBits)))
	// 需要补零的个数
	extraZeros := totalLength - len(refFloats) - len(subFloats)
	// 2 的倍数长度
	power2Len := extraZeros + len(refFloats) + len(subFloats)
	// ----------------------------------------------------------
	// 对于 sub 需要在前面补零
	power2Sub := make([]float64, power2Len)
	fillUpZeroLen4Sub := power2Len - len(subFloats)
	for i := 0; i < fillUpZeroLen4Sub; i++ {
		power2Sub[i] = 0
	}
	for i := 0; i < len(subFloats); i++ {
		power2Sub[fillUpZeroLen4Sub+i] = subFloats[i]
	}
	// 可选择的 FFT 实现 "github.com/brettbuddin/fourier"
	//subFT := fourier.Forward()
	// 先初始化一个 fft 共用实例
	fftIns := fourier.NewFFT(len(power2Sub))
	fftIns.Reset(len(power2Sub))
	subFT := fftIns.Coefficients(nil, power2Sub)
	// ----------------------------------------------------------
	// 对于 ref 需要在后面补零
	power2Ref := make([]float64, power2Len)
	for i := 0; i < len(refFloats); i++ {
		power2Ref[i] = refFloats[i]
	}
	for i := 0; i < power2Len-len(refFloats); i++ {
		power2Ref[len(refFloats)+i] = 0
	}
	// 反转 power2Ref  0, 1，1，0，0 -> 0,0,1,1,0
	for i, j := 0, len(power2Ref)-1; i < j; i, j = i+1, j-1 {
		power2Ref[i], power2Ref[j] = power2Ref[j], power2Ref[i]
	}
	fftIns.Reset(len(power2Ref))
	refFT := fftIns.Coefficients(nil, power2Ref)
	// ----------------------------------------------------------
	// 先计算 subFT * refFT，结果放置在 refFT
	cmplxs.Mul(refFT, subFT)
	// 然后执行 numpy 的 ifft 操作
	convolve := fftIns.Sequence(nil, refFT)
	floats.Scale(1/float64(len(power2Ref)), convolve)

	return convolve
}

func (f FFTAligner) eliminateExtremeOffsetsFromSolutions(convolve, subSting []float64) []float64 {
	if f.maxOffsetSamples == 0 {
		return convolve
	}

	convolveCopy := convolve
	offsetFun := func(offset int) int {
		return len(convolveCopy) - 1 + offset - len(subSting)
	}
	s1 := offsetFun(-f.maxOffsetSamples)
	s2 := offsetFun(f.maxOffsetSamples)
	for i := 0; i < s1; i++ {
		convolveCopy[i] = math.NaN()
	}
	for i := s2; i < len(convolveCopy); i++ {
		convolveCopy[i] = math.NaN()
	}

	return convolveCopy
}

// computeArgmax 找对最优偏移，还需要根据实际情况进行转换（比如，1 步 是 10 ms）
func (f FFTAligner) computeArgmax(convolve, subFloats []float64) (int, float64) {

	convolveTM := treemap.NewWith(utils.Float64Comparator)
	for i, value := range convolve {
		convolveTM.Put(value, i)
	}
	bestScore, bestIndex := convolveTM.Max()

	bestOffset := len(convolve) - 1 - bestIndex.(int) - len(subFloats)

	return bestOffset, bestScore.(float64)
}
