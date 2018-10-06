package threefbx

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg"
	"github.com/oakmound/oak/alg/floatgeom"
)

//TODO consider how they use transformdata and how it creates transforms
//can we skip transform data and just jam things on as we get them?

func generateTransform(td TransformData) mgl64.Mat4 {
	order := ZYXOrder
	if td.eulerOrder != nil {
		order = *td.eulerOrder
	}

	translation := floatgeom.Point3{}
	if td.translation != nil {
		translation = *td.translation
	}

	if td.rotationOffset != nil {
		translation = translation.Add(*td.rotationOffset)
	}

	rotation := mgl64.Ident4()
	if td.rotation != nil {
		rot := td.rotation.MulConst(alg.DegToRad)
		rotation = Euler{rot, order}.makeRotation()
	}

	if td.preRotation != nil {
		rot := td.preRotation.MulConst(alg.DegToRad)
		mat := Euler{rot, order}.makeRotation()
		rotation = mat.Mul4(rotation)
	}
	if td.postRotation != nil {
		rot := td.postRotation.MulConst(alg.DegToRad)
		mat := Euler{rot, order}.makeRotation()
		mat = mat.Inv()
		rotation = rotation.Mul4(mat)
	}

	transform := mgl64.Ident4()

	if td.scale != nil {
		transform = scaleMat4(transform, *td.scale)
	}
	transform = PositionMat4(transform, translation)
	transform = transform.Mul4(rotation)
	return transform
}

func PositionMat4(m mgl64.Mat4, t floatgeom.Point3) mgl64.Mat4 {
	m[12] = t.X()
	m[13] = t.Y()
	m[14] = t.Z()
	return m
}

type InfoObject struct {
	MappingType   string
	ReferenceType string
	DataSize      int
	Indices       []int
	Buffer        []byte
}

// extracts the data from the correct position in the FBX array based on indexing type
func getData(polygonVertexIndex, polygonIndex, vertexIndex int, info InfoObject) []byte {
	var index int
	switch info.MappingType {
	case "ByPolygonVertex":
		index = polygonVertexIndex
	case "ByPolygon":
		index = polygonIndex
	case "ByVertice":
		index = vertexIndex
	case "AllSame":
		index = info.Indices[0]
	default:
		fmt.Println("THREE.FBXLoader: unknown attribute mapping type " + info.MappingType)
	}
	if info.ReferenceType == "IndexToDirect" {
		index = info.Indices[index]
	}
	from := index * info.DataSize
	to := from + info.DataSize
	out := make([]byte, info.DataSize)
	copy(out, info.Buffer[from:to])
	return out
}
