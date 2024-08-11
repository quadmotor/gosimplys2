package main

// type SymetricMatrix struct {
// 	m [10]float64
// }

type SymetricMatrix [10]float64

func planeToQuadric(a float64, b float64, c float64, d float64) SymetricMatrix {

	sym := SymetricMatrix{}

	sym[0] = a * a
	sym[1] = a * b
	sym[2] = a * c
	sym[3] = a * d
	sym[4] = b * b
	sym[5] = b * c
	sym[6] = b * d
	sym[7] = c * c
	sym[8] = c * d
	sym[9] = d * d

	return sym
}

func vertex_error(q *SymetricMatrix, x float64, y float64, z float64) float64 {
	return q[0]*x*x + 2*q[1]*x*y + 2*q[2]*x*z + 2*q[3]*x + q[4]*y*y + 2*q[5]*y*z + 2*q[6]*y + q[7]*z*z + 2*q[8]*z + q[9]
}

func symAdd(a *SymetricMatrix, b *SymetricMatrix) SymetricMatrix {
	c := SymetricMatrix{}

	for i := 0; i < 10; i++ {
		c[i] = a[i] + b[i]
	}

	return c
}
