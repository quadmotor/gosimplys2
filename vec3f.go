package gosimplys2

import "math"

type vec3f [3]float64

var (
	vec3fzero = vec3f{}
)

// a-b
func vec3fminus(a *vec3f, b *vec3f) vec3f {
	r := vec3f{}

	r[0] = a[0] - b[0]
	r[1] = a[1] - b[1]
	r[2] = a[2] - b[2]

	return r
}

func vec3fadd(a *vec3f, b *vec3f) vec3f {
	r := vec3f{}

	r[0] = a[0] + b[0]
	r[1] = a[1] + b[1]
	r[2] = a[2] + b[2]

	return r
}

func vec3fdot(a *vec3f, b *vec3f) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func vec3fnormalize(a *vec3f) {
	l2 := vec3fdot(a, a)
	l2 = math.Sqrt(l2)
	a[0] /= l2
	a[1] /= l2
	a[2] /= l2
}

func vec3fcross(a *vec3f, b *vec3f) vec3f {
	v := vec3f{}

	v[0] = a[1]*b[2] - a[2]*b[1]
	v[1] = a[2]*b[0] - a[0]*b[2]
	v[2] = a[0]*b[1] - a[1]*b[0]

	return v
}
