package sub_timeline_fixer

import (
	"gonum.org/v1/gonum/cmplxs"
	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/floats"
	"math"
)

/*
	复现 https://github.com/smacke/ffsubsync 的算法
*/
type FFTAligner struct {
}

func (f FFTAligner) fit(refFloats, subFloats []float64) {

	// 先初始化一个 fft 共用实例
	fftIns := fourier.NewFFT(1000)
	// 计算出一维矩阵的长度
	total_bits := math.Log2(float64(len(refFloats)) + float64(len(subFloats)))
	total_length := int(math.Pow(2, math.Ceil(total_bits)))
	// 需要补零的个数
	extra_zeros := total_length - len(refFloats) - len(subFloats)
	// 2 的倍数长度
	power2Len := extra_zeros + len(refFloats) + len(subFloats)
	// ----------------------------------------------------------
	// 对于 sub 需要在前面补零
	power2Sub := make([]float64, power2Len)
	for i := 0; i < extra_zeros+len(refFloats); i++ {
		power2Sub[i] = 0
	}
	for i := 0; i < len(subFloats); i++ {
		power2Sub[extra_zeros+len(subFloats)+i] = subFloats[i]
	}
	// "github.com/brettbuddin/fourier"
	//subFT := fourier.Forward()
	fftIns.Reset(len(power2Sub))
	subFT := fftIns.Coefficients(nil, power2Sub)
	// ----------------------------------------------------------
	// 对于 ref 需要在后面补零
	power2Ref := make([]float64, power2Len)
	for i := 0; i < len(refFloats); i++ {
		power2Ref[i] = refFloats[i]
	}
	for i := 0; i < extra_zeros+len(subFloats); i++ {
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
	gotRefFT := fftIns.Sequence(nil, refFT)
	floats.Scale(1/float64(len(power2Ref)), gotRefFT)

	//refFloatsVec := mat.NewVecDense(len(refFloats), refFloats)
	//subFloatsVec := mat.NewVecDense(len(subFloats), subFloats)
	println("d")
	//a := mat.NewVecDense(extra_zeros+refFloatsVec.Len(), nil)
}

func float642comolex() {

}
