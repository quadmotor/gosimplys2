package gosimplys2

import (
	"fmt"
	"math"
	"os"
)

type Vertex struct {
	p      vec3f
	tstart int
	tcount int
	q      SymetricMatrix
	border bool
}

type Triangle struct {
	v       [3]int
	err     [4]float64
	n       vec3f
	deleted bool
	dirty   bool
}

type Ref struct {
	tid     int
	tvertex int
}

type Mesh struct {
	vertices  []Vertex
	triangles []Triangle
	refs      []Ref
}

func (msh *Mesh) calculate_error(id_v1 int, id_v2 int) float64 {
	q := symAdd(&msh.vertices[id_v1].q, &msh.vertices[id_v2].q)
	// border := msh.vertices[id_v1].border && msh.vertices[id_v2].border;
	x := msh.vertices[id_v2].p[0]
	y := msh.vertices[id_v2].p[1]
	z := msh.vertices[id_v2].p[2]
	return vertex_error(&q, x, y, z)
}

// Check if a triangle flips when this edge is removed

func (msh *Mesh) flipped(p vec3f, i0 int, i1 int, v0 *Vertex, v1 *Vertex, deleted *[]int) bool {
	bordercount := 0
	for k := 0; k < v0.tcount; k++ {
		t := &msh.triangles[msh.refs[v0.tstart+k].tid]
		if t.deleted {
			continue
		}

		s := msh.refs[v0.tstart+k].tvertex
		id1 := t.v[(s+1)%3]
		id2 := t.v[(s+2)%3]

		if id1 == i1 || id2 == i1 { // delete ? {
			bordercount++
			(*deleted)[k] = 1
			continue
		}
		d1 := vec3fminus(&msh.vertices[id1].p, &p)
		vec3fnormalize(&d1)
		d2 := vec3fminus(&msh.vertices[id2].p, &p)
		vec3fnormalize(&d2)
		if math.Abs(vec3fdot(&d1, &d2)) > 0.999 {
			return true
		}
		n := vec3f{}
		n = vec3fcross(&d1, &d2)
		vec3fnormalize(&n)
		(*deleted)[k] = 0
		if vec3fdot(&n, &t.n) < 0.2 {
			return true
		}
	}
	return false
}

// Update triangle connections and edge error after a edge is collapsed

func (msh *Mesh) update_triangles(i0 int, v *Vertex, deleted *[]int, deleted_triangles *int) {

	for k := 0; k < v.tcount; k++ {
		r := &msh.refs[v.tstart+k]
		t := &msh.triangles[r.tid]
		if t.deleted {
			continue
		}
		if 1 == (*deleted)[k] {
			t.deleted = true
			*deleted_triangles++
			continue
		}
		t.v[r.tvertex] = i0
		t.dirty = true
		t.err[0] = msh.calculate_error(t.v[0], t.v[1])
		t.err[1] = msh.calculate_error(t.v[1], t.v[2])
		t.err[2] = msh.calculate_error(t.v[2], t.v[0])
		t.err[3] = math.Min(t.err[0], math.Min(t.err[1], t.err[2]))
		msh.refs = append(msh.refs, *r) // msh.refs.push_back(r);
	}
}

func resizeslice(refs *[]int, n3 int) {

	if cap(*refs) > n3 {
		*refs = (*refs)[:n3]
	} else {
		*refs = make([]int, n3)
	}
}

func (msh *Mesh) writeObj(output string) {
	f, _ := os.Create(output)
	defer f.Close()

	for i := 0; i < len(msh.vertices); i++ {
		fmt.Fprintf(f, "v %g %g %g\n",
			msh.vertices[i].p[0], msh.vertices[i].p[1], msh.vertices[i].p[2])
	}

	for i := 0; i < len(msh.triangles); i++ {
		if !msh.triangles[i].deleted {
			fmt.Fprintf(f, "f %d// %d// %d//\n",
				msh.triangles[i].v[0]+1, msh.triangles[i].v[1]+1, msh.triangles[i].v[2]+1)
		}
	}
}

// compact triangles, compute edge error and build reference list

func (msh *Mesh) update_mesh(iteration int) {
	if iteration > 0 { // compact triangles
		dst := 0
		for i := 0; i < len(msh.triangles); i++ {
			if !msh.triangles[i].deleted {
				msh.triangles[dst] = msh.triangles[i]
				dst++
			}
		}
		msh.triangles = msh.triangles[:dst]
	}
	//

	// Init Reference ID list
	// loopi(0,vertices.size())
	for i := 0; i < len(msh.vertices); i++ {
		msh.vertices[i].tstart = 0
		msh.vertices[i].tcount = 0
	}
	// loopi(0,triangles.size()){
	for i := 0; i < len(msh.triangles); i++ {
		t := &msh.triangles[i]
		for j := 0; j < 3; j++ {
			msh.vertices[t.v[j]].tcount++
		}
	}
	tstart := 0
	for i := 0; i < len(msh.vertices); i++ {
		v := &msh.vertices[i]
		v.tstart = tstart
		tstart += v.tcount
		v.tcount = 0
	}

	// Write References
	n3 := len(msh.triangles) * 3
	if cap(msh.refs) > n3 {
		msh.refs = msh.refs[:n3]
	} else {
		msh.refs = make([]Ref, n3)
	}

	for i := 0; i < len(msh.triangles); i++ {
		t := &msh.triangles[i]
		for j := 0; j < 3; j++ {
			v := &msh.vertices[t.v[j]]
			msh.refs[v.tstart+v.tcount].tid = i
			msh.refs[v.tstart+v.tcount].tvertex = j
			v.tcount++
		}
	}

	// Init Quadrics by Plane & Edge Errors
	//
	// required at the beginning ( iteration == 0 )
	// recomputing during the simplification is not required,
	// but mostly improves the result for closed meshes
	//
	if iteration == 0 {
		// Identify boundary : vertices[].border=0,1

		vcount := []int{}
		vids := []int{}

		for i := 0; i < len(msh.vertices); i++ {
			msh.vertices[i].border = false
		}

		for i := 0; i < len(msh.vertices); i++ {
			v := &msh.vertices[i]
			vcount = vcount[:0]
			vids = vids[:0]
			for j := 0; j < v.tcount; j++ {
				k := msh.refs[v.tstart+j].tid
				t := &msh.triangles[k]
				for k := 0; k < 3; k++ {
					ofs := 0
					id := t.v[k]
					//while(ofs<vcount.size())
					for {
						if ofs >= len(vcount) {
							break
						}
						if vids[ofs] == id {
							break
						}
						ofs++
					}
					if ofs == len(vcount) {
						vcount = append(vcount, 1)
						vids = append(vids, id)
					} else {
						vcount[ofs]++
					}
				}
			}
			for j := 0; j < len(vcount); j++ {
				if vcount[j] == 1 {
					msh.vertices[vids[j]].border = true
				}
			}
		}
		//initialize errors
		for i := 0; i < len(msh.vertices); i++ {
			msh.vertices[i].q = SymetricMatrix{} //(0.0)
		}

		for i := 0; i < len(msh.triangles); i++ {

			t := &msh.triangles[i]
			n := vec3f{}
			p := [3]vec3f{}
			for j := 0; j < 3; j++ {
				p[j] = msh.vertices[t.v[j]].p
			}

			d01 := vec3fminus(&p[1], &p[0])
			d02 := vec3fminus(&p[2], &p[0])
			n = vec3fcross(&d01, &d02)
			vec3fnormalize(&n)
			t.n = n

			dis := -vec3fdot(&n, &p[0])
			q0 := planeToQuadric(n[0], n[1], n[2], dis)
			for j := 0; j < 3; j++ {
				msh.vertices[t.v[j]].q = symAdd(&msh.vertices[t.v[j]].q, &q0)
			}
		}

		for i := 0; i < len(msh.triangles); i++ {
			// Calc Edge Error
			t := &msh.triangles[i]
			// p := vec3f{}
			for j := 0; j < 3; j++ {
				t.err[j] = msh.calculate_error(t.v[j], t.v[(j+1)%3])
			}
			t.err[3] = math.Min(t.err[0], math.Min(t.err[1], t.err[2]))
		}
	}
}

// Finally compact mesh before exiting

func (msh *Mesh) compact_mesh() {

	dst := 0

	for i := 0; i < len(msh.vertices); i++ {
		msh.vertices[i].tcount = 0
	}

	for i := 0; i < len(msh.triangles); i++ {
		if !msh.triangles[i].deleted {
			t := &msh.triangles[i]
			msh.triangles[dst] = *t
			dst++
			for j := 0; j < 3; j++ {
				msh.vertices[t.v[j]].tcount = 1
			}
		}
	}
	msh.triangles = msh.triangles[:dst]
	dst = 0

	for i := 0; i < len(msh.vertices); i++ {
		if msh.vertices[i].tcount > 0 {
			msh.vertices[i].tstart = dst
			msh.vertices[dst].p = msh.vertices[i].p
			dst++
		}
	}

	for i := 0; i < len(msh.triangles); i++ {
		t := &msh.triangles[i]
		for j := 0; j < 3; j++ {
			t.v[j] = msh.vertices[t.v[j]].tstart
		}
	}
	msh.vertices = msh.vertices[:dst]
}

// Main simplification function
//
// target_count  : target nr. of triangles
// agressiveness : sharpness to increase the threshold.
//
//	5..8 are good numbers
//	more iterations yield higher quality
//
// =7
func (msh *Mesh) simplify_mesh(target_count int, agressiveness float64) {
	// init
	// printf("%s - start\n",__FUNCTION__);
	// int timeStart=timeGetTime();

	for i := 0; i < len(msh.triangles); i++ {
		msh.triangles[i].deleted = false
	}

	// main iteration loop

	deleted_triangles := 0
	deleted0 := []int{}
	deleted1 := []int{}
	triangle_count := len(msh.triangles)

	for iteration := 0; iteration < 100; iteration++ {

		// target number of triangles reached ? Then break
		// printf("iteration %d - triangles %d\n",iteration,triangle_count-deleted_triangles);
		if triangle_count-deleted_triangles <= target_count {
			break
		}

		// update mesh once in a while
		if iteration%5 == 0 {
			msh.update_mesh(iteration)
		}

		// clear dirty flag
		for i := 0; i < len(msh.triangles); i++ {
			msh.triangles[i].dirty = false
		}
		//
		// All triangles with edges below the threshold will be removed
		//
		// The following numbers works well for most models.
		// If it does not, try to adjust the 3 parameters
		//
		threshold := 0.000000001 * math.Pow(float64(iteration+3), agressiveness)

		// remove vertices & mark deleted triangles
		for i := 0; i < len(msh.triangles); i++ {
			t := &msh.triangles[i]
			if t.err[3] > threshold {
				continue
			}
			if t.deleted {
				continue
			}
			if t.dirty {
				continue
			}

			for j := 0; j < 3; j++ {
				if t.err[j] < threshold {
					i0 := t.v[j]
					v0 := &msh.vertices[i0]
					i1 := t.v[(j+1)%3]
					v1 := &msh.vertices[i1]

					// Border check
					if v0.border != v1.border {
						continue
					}

					// Compute vertex to collapse to
					p := v1.p
					msh.calculate_error(i0, i1)

					resizeslice(&deleted0, v0.tcount)
					resizeslice(&deleted1, v1.tcount)

					// don't remove if flipped
					if msh.flipped(p, i0, i1, v0, v1, &deleted0) {
						continue
					}
					if msh.flipped(p, i1, i0, v1, v0, &deleted1) {
						continue
					}

					// not flipped, so remove edge
					v0.p = p
					v0.q = symAdd(&v1.q, &v0.q)
					tstart := len(msh.refs)

					msh.update_triangles(i0, v0, &deleted0, &deleted_triangles)
					msh.update_triangles(i0, v1, &deleted1, &deleted_triangles)

					tcount := len(msh.refs) - tstart

					if tcount <= v0.tcount {
						// save ram
						if tcount > 0 {
							for k := 0; k < tcount; k++ {
								msh.refs[v0.tstart+k] = msh.refs[tstart+k]
							}
						}
					} else {
						// append
						v0.tstart = tstart
					}

					v0.tcount = tcount
					break
				}
			}
			// done?
			if triangle_count-deleted_triangles <= target_count {
				break
			}
		}

	}

	// clean up mesh
	msh.compact_mesh()

}
