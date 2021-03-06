package gltf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// An Asset is metadata about the glTF asset.
type Asset struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Copyright  string      `json:"copyright,omitempty"`         // A copyright message suitable for display to credit the content creator.
	Generator  string      `json:"generator,omitempty"`         // Tool that generated this glTF model. Useful for debugging.
	Version    string      `json:"version" validate:"required"` // The glTF version that this asset targets.
	MinVersion string      `json:"minVersion,omitempty"`        // The minimum glTF version that this asset targets.
}

// Document defines the root object for a glTF asset.
type Document struct {
	Extensions         Extensions   `json:"extensions,omitempty"`
	Extras             interface{}  `json:"extras,omitempty"`
	ExtensionsUsed     []string     `json:"extensionsUsed,omitempty"`
	ExtensionsRequired []string     `json:"extensionsRequired,omitempty"`
	Accessors          []Accessor   `json:"accessors,omitempty" validate:"dive"`
	Animations         []Animation  `json:"animations,omitempty" validate:"dive"`
	Asset              Asset        `json:"asset"`
	Buffers            []Buffer     `json:"buffers,omitempty" validate:"dive"`
	BufferViews        []BufferView `json:"bufferViews,omitempty" validate:"dive"`
	Cameras            []Camera     `json:"cameras,omitempty" validate:"dive"`
	Images             []Image      `json:"images,omitempty" validate:"dive"`
	Materials          []Material   `json:"materials,omitempty" validate:"dive"`
	Meshes             []Mesh       `json:"meshes,omitempty" validate:"dive"`
	Nodes              []Node       `json:"nodes,omitempty" validate:"dive"`
	Samplers           []Sampler    `json:"samplers,omitempty" validate:"dive"`
	Scene              int32        `json:"scene" validate:"gte=-1"`
	Scenes             []Scene      `json:"scenes,omitempty" validate:"dive"`
	Skins              []Skin       `json:"skins,omitempty" validate:"dive"`
	Textures           []Texture    `json:"textures,omitempty" validate:"dive"`
}

// UnmarshalJSON unmarshal the document with the correct default values.
func (d *Document) UnmarshalJSON(data []byte) error {
	type alias Document
	tmp := &alias{Scene: -1}
	err := json.Unmarshal(data, tmp)
	if err == nil {
		*d = Document(*tmp)
	}
	return err
}

// MarshalJSON marshal the document with the correct default values.
func (d *Document) MarshalJSON() ([]byte, error) {
	type alias Document
	if d.Scene == -1 {
		return json.Marshal(&struct {
			Scene int32 `json:"scene,omitempty"`
			*alias
		}{
			Scene: 0,
			alias: (*alias)(d),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(d)})
}

// An Accessor is a typed view into a bufferView.
// An accessor provides a typed view into a bufferView or a subset of a bufferView
// similar to how WebGL's vertexAttribPointer() defines an attribute in a buffer.
type Accessor struct {
	Extensions    Extensions    `json:"extensions,omitempty"`
	Extras        interface{}   `json:"extras,omitempty"`
	Name          string        `json:"name,omitempty"`
	BufferView    int32         `json:"bufferView" validate:"gte=-1"`
	ByteOffset    uint32        `json:"byteOffset,omitempty"`
	ComponentType ComponentType `json:"componentType" validate:"oneof=5120 5121 5122 5123 5125 5126"`
	Normalized    bool          `json:"normalized,omitempty"`      // Specifies whether integer data values should be normalized.
	Count         uint32        `json:"count" validate:"required"` // The number of attributes referenced by this accessor.
	Type          AccessorType  `json:"type" validate:"oneof=SCALAR VEC2 VEC3 VEC4 MAT2 MAT3 MAT4"`
	Max           []float64     `json:"max,omitempty" validate:"omitempty,lte=16"` // Maximum value of each component in this attribute.
	Min           []float64     `json:"min,omitempty" validate:"omitempty,lte=16"` // Minimum value of each component in this attribute.
	Sparse        *Sparse       `json:"sparse,omitempty"`                          // Sparse storage of attributes that deviate from their initialization value.
}

// NewAccessor returns a default accessor.
func NewAccessor() *Accessor {
	return &Accessor{BufferView: -1}
}

// UnmarshalJSON unmarshal the accessor with the correct default values.
func (a *Accessor) UnmarshalJSON(data []byte) error {
	type alias Accessor
	tmp := alias(*NewAccessor())
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*a = Accessor(tmp)
	}
	return err
}

// MarshalJSON marshal the accessor with the correct default values.
func (a *Accessor) MarshalJSON() ([]byte, error) {
	type alias Accessor
	if a.BufferView == -1 {
		return json.Marshal(&struct {
			BufferView int32 `json:"bufferView,omitempty"`
			*alias
		}{
			BufferView: 0,
			alias:      (*alias)(a),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(a)})
}

// Sparse storage of attributes that deviate from their initialization value.
type Sparse struct {
	Extensions Extensions    `json:"extensions,omitempty"`
	Extras     interface{}   `json:"extras,omitempty"`
	Count      uint32        `json:"count" validate:"gte=1"` // Number of entries stored in the sparse array.
	Indices    SparseIndices `json:"indices"`                // Index array of size count that points to those accessor attributes that deviate from their initialization value.
	Values     SparseValues  `json:"values"`                 // Array of size count times number of components, storing the displaced accessor attributes pointed by indices.
}

// SparseValues stores the displaced accessor attributes pointed by accessor.sparse.indices.
type SparseValues struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	BufferView uint32      `json:"bufferView"`
	ByteOffset uint32      `json:"byteOffset,omitempty"`
}

// SparseIndices defines the indices of those attributes that deviate from their initialization value.
type SparseIndices struct {
	Extensions    Extensions    `json:"extensions,omitempty"`
	Extras        interface{}   `json:"extras,omitempty"`
	BufferView    uint32        `json:"bufferView"`
	ByteOffset    uint32        `json:"byteOffset,omitempty"`
	ComponentType ComponentType `json:"componentType" validate:"oneof=5121 5123 5125"`
}

// A Buffer points to binary geometry, animation, or skins.
type Buffer struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	URI        string      `json:"uri,omitempty" validate:"omitempty"`
	ByteLength uint32      `json:"byteLength" validate:"required"`
	Data       []uint8     `json:"-"`
}

// IsEmbeddedResource returns true if the buffer points to an embedded resource.
func (b *Buffer) IsEmbeddedResource() bool {
	return strings.HasPrefix(b.URI, mimetypeApplicationOctet)
}

// EmbeddedResource defines the buffer as an embedded resource and encodes the URI so it points to the the resource.
func (b *Buffer) EmbeddedResource() {
	b.URI = fmt.Sprintf("%s,%s", mimetypeApplicationOctet, base64.StdEncoding.EncodeToString(b.Data))
}

// marshalData decode the buffer from the URI. If the buffer is not en embedded resource the returned array will be empty.
func (b *Buffer) marshalData() ([]uint8, error) {
	if !b.IsEmbeddedResource() {
		return []uint8{}, nil
	}
	startPos := len(mimetypeApplicationOctet) + 1
	return base64.StdEncoding.DecodeString(b.URI[startPos:])
}

// BufferView is a view into a buffer generally representing a subset of the buffer.
type BufferView struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Buffer     int32       `json:"buffer" validate:"gte=-1"`
	ByteOffset uint32      `json:"byteOffset,omitempty"`
	ByteLength uint32      `json:"byteLength" validate:"required"`
	ByteStride uint32      `json:"byteStride,omitempty" validate:"omitempty,gte=4,lte=252"`
	Target     Target      `json:"target,omitempty" validate:"omitempty,oneof=34962 34963"`
}

// UnmarshalJSON unmarshal the buffer view with the correct default values.
func (b *BufferView) UnmarshalJSON(data []byte) error {
	type alias BufferView
	tmp := &alias{Buffer: -1}
	err := json.Unmarshal(data, tmp)
	if err == nil {
		*b = BufferView(*tmp)
	}
	return err
}

// MarshalJSON marshal the buffer view with the correct default values.
func (b *BufferView) MarshalJSON() ([]byte, error) {
	type alias BufferView
	if b.Buffer == -1 {
		return json.Marshal(&struct {
			Buffer int32 `json:"buffer,omitempty"`
			*alias
		}{
			Buffer: 0,
			alias:  (*alias)(b),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(b)})
}

// The Scene contains a list of root nodes.
type Scene struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	Nodes      []uint32    `json:"nodes,omitempty" validate:"omitempty,unique"`
}

// A Node in the node hierarchy.
// It can have either a matrix or any combination of translation/rotation/scale (TRS) properties.
type Node struct {
	Extensions  Extensions  `json:"extensions,omitempty"`
	Extras      interface{} `json:"extras,omitempty"`
	Name        string      `json:"name,omitempty"`
	Camera      int32       `json:"camera" validate:"gte=-1"`
	Children    []uint32    `json:"children,omitempty" validate:"omitempty,unique"`
	Skin        int32       `json:"skin" validate:"gte=-1"`
	Matrix      [16]float64 `json:"matrix"` // A 4x4 transformation matrix stored in column-major order.
	Mesh        int32       `json:"mesh" validate:"gte=-1"`
	Rotation    [4]float64  `json:"rotation" validate:"omitempty,dive,gte=-1,lte=1"` // The node's unit quaternion rotation in the order (x, y, z, w), where w is the scalar.
	Scale       [3]float64  `json:"scale"`
	Translation [3]float64  `json:"translation"`
	Weights     []float64   `json:"weights,omitempty"` // The weights of the instantiated Morph Target.
}

// NewNode returns a default Node.
func NewNode() *Node {
	return &Node{
		Matrix:   [16]float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1},
		Rotation: [4]float64{0, 0, 0, 1},
		Scale:    [3]float64{1, 1, 1},
		Camera:   -1,
		Skin:     -1,
		Mesh:     -1,
	}
}

// UnmarshalJSON unmarshal the node with the correct default values.
func (n *Node) UnmarshalJSON(data []byte) error {
	type alias Node
	tmp := alias(*NewNode())
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*n = Node(tmp)
	}
	return err
}

// MarshalJSON marshal the node with the correct default values.
func (n *Node) MarshalJSON() ([]byte, error) {
	type alias Node
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(n)})
	if err == nil {
		if n.Matrix == [16]float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1} {
			out = removeProperty([]byte(`"matrix":[1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1]`), out)
		}
		if n.Rotation == [4]float64{0, 0, 0, 1} {
			out = removeProperty([]byte(`"rotation":[0,0,0,1]`), out)
		}
		if n.Scale == [3]float64{1, 1, 1} {
			out = removeProperty([]byte(`"scale":[1,1,1]`), out)
		}
		if n.Translation == [3]float64{0, 0, 0} {
			out = removeProperty([]byte(`"translation":[0,0,0]`), out)
		}
		if n.Camera == -1 {
			out = removeProperty([]byte(`"camera":-1`), out)
		}
		if n.Skin == -1 {
			out = removeProperty([]byte(`"skin":-1`), out)
		}
		if n.Mesh == -1 {
			out = removeProperty([]byte(`"mesh":-1`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// Skin defines joints and matrices.
type Skin struct {
	Extensions          Extensions  `json:"extensions,omitempty"`
	Extras              interface{} `json:"extras,omitempty"`
	Name                string      `json:"name,omitempty"`
	InverseBindMatrices int32       `json:"inverseBindMatrices" validate:"gte=-1"` // The index of the accessor containing the floating-point 4x4 inverse-bind matrices.
	Skeleton            int32       `json:"skeleton" validate:"gte=-1"`            // The index of the node used as a skeleton root. When undefined, joints transforms resolve to scene root.
	Joints              []uint32    `json:"joints" validate:"omitempty,unique"`    // Indices of skeleton nodes, used as joints in this skin.
}

// NewSkin create a default Skin.
func NewSkin() *Skin {
	return &Skin{InverseBindMatrices: -1, Skeleton: -1}
}

// UnmarshalJSON unmarshal the skin with the correct default values.
func (s *Skin) UnmarshalJSON(data []byte) error {
	type alias Skin
	tmp := alias(*NewSkin())
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*s = Skin(tmp)
	}
	return err
}

// MarshalJSON marshal the skin with the correct default values.
func (s *Skin) MarshalJSON() ([]byte, error) {
	type alias Skin
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(s)})
	if err == nil {
		if s.InverseBindMatrices == -1 {
			out = removeProperty([]byte(`"inverseBindMatrices":-1`), out)
		}
		if s.Skeleton == -1 {
			out = removeProperty([]byte(`"skeleton":-1`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// A Camera projection. A node can reference a camera to apply a transform to place the camera in the scene.
type Camera struct {
	Extensions   Extensions    `json:"extensions,omitempty"`
	Extras       interface{}   `json:"extras,omitempty"`
	Name         string        `json:"name,omitempty"`
	Orthographic *Orthographic `json:"orthographic,omitempty"`
	Perspective  *Perspective  `json:"perspective,omitempty"`
	Type         CameraType    `json:"type" validate:"oneof=perspective orthographic"`
}

// Orthographic camera containing properties to create an orthographic projection matrix.
type Orthographic struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Xmag       float64     `json:"xmag"`                               // The horizontal magnification of the view.
	Ymag       float64     `json:"ymag"`                               // The vertical magnification of the view.
	Zfar       float64     `json:"zfar" validate:"gt=0,gtfield=Znear"` // The distance to the far clipping plane.
	Znear      float64     `json:"znear" validate:"gte=0"`             // The distance to the near clipping plane.
}

// Perspective camera containing properties to create a perspective projection matrix.
type Perspective struct {
	Extensions  Extensions  `json:"extensions,omitempty"`
	Extras      interface{} `json:"extras,omitempty"`
	AspectRatio float64     `json:"aspectRatio,omitempty"`
	Yfov        float64     `json:"yfov"`           // The vertical field of view in radians.
	Zfar        float64     `json:"zfar,omitempty"` // The distance to the far clipping plane.
	Znear       float64     `json:"znear"`          // The distance to the near clipping plane.
}

// A Mesh is a set of primitives to be rendered. A node can contain one mesh. A node's transform places the mesh in the scene.
type Mesh struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	Primitives []Primitive `json:"primitives" validate:"required,gt=0,dive"`
	Weights    []float64   `json:"weights,omitempty"`
}

// Primitive defines the geometry to be rendered with the given material.
type Primitive struct {
	Extensions Extensions    `json:"extensions,omitempty"`
	Extras     interface{}   `json:"extras,omitempty"`
	Attributes Attribute     `json:"attributes"`
	Indices    int32         `json:"indices" validate:"gte=-1"` // The index of the accessor that contains the indices.
	Material   int32         `json:"material" validate:"gte=-1"`
	Mode       PrimitiveMode `json:"mode" validate:"lte=6"`
	Targets    []Attribute   `json:"targets,omitempty" validate:"omitempty,dive,dive,keys,oneof=POSITION NORMAL TANGENT,endkeys"` // Only POSITION, NORMAL, and TANGENT supported.
}

// NewPrimitive create a default Primitive.
func NewPrimitive() *Primitive {
	return &Primitive{Mode: Triangles, Indices: -1, Material: -1}
}

// UnmarshalJSON unmarshal the primitive with the correct default values.
func (p *Primitive) UnmarshalJSON(data []byte) error {
	type alias Primitive
	tmp := alias(*NewPrimitive())
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*p = Primitive(tmp)
	}
	return err
}

// MarshalJSON marshal the primitive with the correct default values.
func (p *Primitive) MarshalJSON() ([]byte, error) {
	type alias Primitive
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(p)})
	if err == nil {
		if p.Indices == -1 {
			out = removeProperty([]byte(`"indices":-1`), out)
		}
		if p.Material == -1 {
			out = removeProperty([]byte(`"material":-1`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// The Material appearance of a primitive.
type Material struct {
	Extensions           Extensions            `json:"extensions,omitempty"`
	Extras               interface{}           `json:"extras,omitempty"`
	Name                 string                `json:"name,omitempty"`
	PBRMetallicRoughness *PBRMetallicRoughness `json:"pbrMetallicRoughness,omitempty"`
	NormalTexture        *NormalTexture        `json:"normalTexture,omitempty"`
	OcclusionTexture     *OcclusionTexture     `json:"occlusionTexture,omitempty"`
	EmissiveTexture      *TextureInfo          `json:"emissiveTexture,omitempty"`
	EmissiveFactor       [3]float64            `json:"emissiveFactor,omitempty" validate:"dive,gte=0,lte=1"`
	AlphaMode            AlphaMode             `json:"alphaMode,omitempty" validate:"oneof=OPAQUE MASK BLEND"`
	AlphaCutoff          float64               `json:"alphaCutoff" validate:"gte=0"`
	DoubleSided          bool                  `json:"doubleSided,omitempty"`
}

// NewMaterial create a default Material.
func NewMaterial() *Material {
	return &Material{AlphaCutoff: 0.5, AlphaMode: Opaque}
}

// UnmarshalJSON unmarshal the material with the correct default values.
func (m *Material) UnmarshalJSON(data []byte) error {
	type alias Material
	tmp := alias(*NewMaterial())
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = Material(tmp)
	}
	return err
}

// MarshalJSON marshal the material with the correct default values.
func (m *Material) MarshalJSON() ([]byte, error) {
	type alias Material
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.AlphaCutoff == 0.5 {
			out = removeProperty([]byte(`"alphaCutoff":0.5`), out)
		}
		if m.AlphaMode == Opaque {
			out = removeProperty([]byte(`"alphaMode":"OPAQUE"`), out)
		}
		if m.EmissiveFactor == [3]float64{0, 0, 0} {
			out = removeProperty([]byte(`"emissiveFactor":[0,0,0]`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// A NormalTexture references to a normal texture.
type NormalTexture struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Index      int32       `json:"index" validate:"gte=-1"`
	TexCoord   uint32      `json:"texCoord,omitempty"` // The index of texture's TEXCOORD attribute used for texture coordinate mapping.
	Scale      float64     `json:"scale"`
}

// NewNormalTexture returns a default NormalTexture.
func NewNormalTexture(index int32) *NormalTexture {
	return &NormalTexture{Index: index, Scale: 1}
}

// UnmarshalJSON unmarshal the texture info with the correct default values.
func (n *NormalTexture) UnmarshalJSON(data []byte) error {
	type alias NormalTexture
	tmp := alias(*NewNormalTexture(-1))
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*n = NormalTexture(tmp)
	}
	return err
}

// MarshalJSON marshal the texture info with the correct default values.
func (n *NormalTexture) MarshalJSON() ([]byte, error) {
	type alias NormalTexture
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(n)})
	if err == nil {
		if n.Index == -1 {
			out = removeProperty([]byte(`"index":-1`), out)
		}
		if n.Scale == -1 {
			out = removeProperty([]byte(`"scale":-1`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// An OcclusionTexture references to an occlusion texture
type OcclusionTexture struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Index      int32       `json:"index" validate:"gte=-1"`
	TexCoord   uint32      `json:"texCoord,omitempty"` // The index of texture's TEXCOORD attribute used for texture coordinate mapping.
	Strength   float64     `json:"strength" validate:"gte=0,lte=1"`
}

// NewOcclusionTexture returns a default OcclusionTexture.
func NewOcclusionTexture(index int32) *OcclusionTexture {
	return &OcclusionTexture{Index: index, Strength: 1}
}

// UnmarshalJSON unmarshal the texture info with the correct default values.
func (o *OcclusionTexture) UnmarshalJSON(data []byte) error {
	type alias OcclusionTexture
	tmp := alias(*NewOcclusionTexture(-1))
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*o = OcclusionTexture(tmp)
	}
	return err
}

// MarshalJSON marshal the texture info with the correct default values.
func (o *OcclusionTexture) MarshalJSON() ([]byte, error) {
	type alias OcclusionTexture
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(o)})
	if err == nil {
		if o.Index == -1 {
			out = removeProperty([]byte(`"index":-1`), out)
		}
		if o.Strength == 1 {
			out = removeProperty([]byte(`"strength":1`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// PBRMetallicRoughness defines a set of parameter values that are used to define the metallic-roughness material model from Physically-Based Rendering (PBR) methodology.
type PBRMetallicRoughness struct {
	Extensions               Extensions   `json:"extensions,omitempty"`
	Extras                   interface{}  `json:"extras,omitempty"`
	BaseColorFactor          [4]float64   `json:"baseColorFactor" validate:"dive,gte=0,lte=1"`
	BaseColorTexture         *TextureInfo `json:"baseColorTexture,omitempty"`
	MetallicFactor           float64      `json:"metallicFactor" validate:"gte=0,lte=1"`
	RoughnessFactor          float64      `json:"roughnessFactor" validate:"gte=0,lte=1"`
	MetallicRoughnessTexture *TextureInfo `json:"metallicRoughnessTexture,omitempty"`
}

// NewPBRMetallicRoughness returns a default PBRMetallicRoughness.
func NewPBRMetallicRoughness() *PBRMetallicRoughness {
	return &PBRMetallicRoughness{BaseColorFactor: [4]float64{1, 1, 1, 1}, MetallicFactor: 1, RoughnessFactor: 1}
}

// UnmarshalJSON unmarshal the pbr with the correct default values.
func (p *PBRMetallicRoughness) UnmarshalJSON(data []byte) error {
	type alias PBRMetallicRoughness
	tmp := alias(*NewPBRMetallicRoughness())
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*p = PBRMetallicRoughness(tmp)
	}
	return err
}

// MarshalJSON marshal the pbr with the correct default values.
func (p *PBRMetallicRoughness) MarshalJSON() ([]byte, error) {
	type alias PBRMetallicRoughness
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(p)})
	if err == nil {
		if p.MetallicFactor == 1 {
			out = removeProperty([]byte(`"metallicFactor":1`), out)
		}
		if p.RoughnessFactor == 1 {
			out = removeProperty([]byte(`"roughnessFactor":1`), out)
		}
		if p.BaseColorFactor == [4]float64{1, 1, 1, 1} {
			out = removeProperty([]byte(`"baseColorFactor":[1,1,1,1]`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// TextureInfo references to a texture.
type TextureInfo struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Index      int32       `json:"index" validate:"gte=-1"`
	TexCoord   uint32      `json:"texCoord,omitempty"` // The index of texture's TEXCOORD attribute used for texture coordinate mapping.
}

// NewTextureInfo returns a default TextureInfo.
func NewTextureInfo(index int32) *TextureInfo {
	return &TextureInfo{Index: index}
}

// UnmarshalJSON unmarshal the texture info with the correct default values.
func (t *TextureInfo) UnmarshalJSON(data []byte) error {
	type alias TextureInfo
	tmp := alias(*NewTextureInfo(-1))
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*t = TextureInfo(tmp)
	}
	return err
}

// MarshalJSON marshal the texture info with the correct default values.
func (t *TextureInfo) MarshalJSON() ([]byte, error) {
	type alias TextureInfo
	if t.Index == -1 {
		return json.Marshal(&struct {
			Index int32 `json:"index,omitempty"`
			*alias
		}{
			Index: 0,
			alias: (*alias)(t),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(t)})
}

// A Texture and its sampler.
type Texture struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	Sampler    int32       `json:"sampler" validate:"gte=-1"`
	Source     int32       `json:"source" validate:"gte=-1"`
}

// NewTexture returns a default Texture.
func NewTexture() *Texture {
	return &Texture{Sampler: -1, Source: -1}
}

// UnmarshalJSON unmarshal the texture with the correct default values.
func (t *Texture) UnmarshalJSON(data []byte) error {
	type alias Texture
	tmp := alias(*NewTexture())
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*t = Texture(tmp)
	}
	return err
}

// MarshalJSON marshal the texture with the correct default values.
func (t *Texture) MarshalJSON() ([]byte, error) {
	type alias Texture
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(t)})
	if err == nil {
		if t.Sampler == -1 {
			out = removeProperty([]byte(`"sampler":-1`), out)
		}
		if t.Source == -1 {
			out = removeProperty([]byte(`"source":-1`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// Sampler of a texture for filtering and wrapping modes.
type Sampler struct {
	Extensions Extensions   `json:"extensions,omitempty"`
	Extras     interface{}  `json:"extras,omitempty"`
	Name       string       `json:"name,omitempty"`
	MagFilter  MagFilter    `json:"magFilter,omitempty" validate:"omitempty,oneof=9728 9729"`
	MinFilter  MinFilter    `json:"minFilter,omitempty" validate:"omitempty,oneof=9728 9729 9984 9985 9986 9987"`
	WrapS      WrappingMode `json:"wrapS,omitempty" validate:"omitempty,oneof=33071 33648 10497"`
	WrapT      WrappingMode `json:"wrapT,omitempty" validate:"omitempty,oneof=33071 33648 10497"`
}

// NewSampler returns a default Sampler.
func NewSampler() *Sampler {
	return &Sampler{WrapS: Repeat, WrapT: Repeat}
}

// UnmarshalJSON unmarshal the sampler with the correct default values.
func (s *Sampler) UnmarshalJSON(data []byte) error {
	type alias Sampler
	tmp := alias(*NewSampler())
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*s = Sampler(tmp)
	}
	return err
}

// Image data used to create a texture. Image can be referenced by URI or bufferView index.
// mimeType is required in the latter case.
type Image struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	URI        string      `json:"uri,omitempty" validate:"omitempty"`
	MimeType   string      `json:"mimeType,omitempty" validate:"omitempty,oneof=image/jpeg image/png"` // Manadatory if BufferView is defined.
	BufferView uint32      `json:"bufferView,omitempty"`                                               // Use this instead of the image's uri property.
}

// IsEmbeddedResource returns true if the buffer points to an embedded resource.
func (im *Image) IsEmbeddedResource() bool {
	return strings.HasPrefix(im.URI, mimetypeImagePNG) || strings.HasPrefix(im.URI, mimetypeImageJPG)
}

// MarshalData decode the image from the URI. If the image is not en embedded resource the returned array will be empty.
func (im *Image) MarshalData() ([]uint8, error) {
	if !im.IsEmbeddedResource() {
		return []uint8{}, nil
	}
	mimetype := mimetypeImagePNG
	if strings.HasPrefix(im.URI, mimetypeImageJPG) {
		mimetype = mimetypeImageJPG
	}
	startPos := len(mimetype) + 1
	return base64.StdEncoding.DecodeString(im.URI[startPos:])
}

// An Animation keyframe.
type Animation struct {
	Extensions Extensions         `json:"extensions,omitempty"`
	Extras     interface{}        `json:"extras,omitempty"`
	Name       string             `json:"name,omitempty"`
	Channels   []Channel          `json:"channels" validate:"required,gt=0,dive"`
	Samplers   []AnimationSampler `json:"samplers" validate:"required,gt=0,dive"`
}

// AnimationSampler combines input and output accessors with an interpolation algorithm to define a keyframe graph (but not its target).
type AnimationSampler struct {
	Extensions    Extensions    `json:"extensions,omitempty"`
	Extras        interface{}   `json:"extras,omitempty"`
	Input         int32         `json:"input" validate:"gte=-1"` // The index of an accessor containing keyframe input values.
	Interpolation Interpolation `json:"interpolation,omitempty" validate:"omitempty,oneof=LINEAR STEP CUBICSPLINE"`
	Output        int32         `json:"output" validate:"gte=-1"` // The index of an accessor containing keyframe output values.
}

// NewAnimationSampler returns a default AnimationSampler.
func NewAnimationSampler() *AnimationSampler {
	return &AnimationSampler{Input: -1, Interpolation: Linear, Output: -1}
}

// UnmarshalJSON unmarshal the animation sampler with the correct default values.
func (as *AnimationSampler) UnmarshalJSON(data []byte) error {
	type alias AnimationSampler
	tmp := alias(*NewAnimationSampler())
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*as = AnimationSampler(tmp)
	}
	return err
}

// The Channel targets an animation's sampler at a node's property.
type Channel struct {
	Extensions Extensions    `json:"extensions,omitempty"`
	Extras     interface{}   `json:"extras,omitempty"`
	Sampler    int32         `json:"sampler" validate:"gte=-1"`
	Target     ChannelTarget `json:"target"`
}

// NewChannel returns a default Channel.
func NewChannel(sampler int32) *Channel {
	return &Channel{Sampler: sampler}
}

// UnmarshalJSON unmarshal the channel with the correct default values.
func (ch *Channel) UnmarshalJSON(data []byte) error {
	type alias Channel
	tmp := alias(*NewChannel(-1))
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*ch = Channel(tmp)
	}
	return err
}

// MarshalJSON marshal the channel with the correct default values.
func (ch *Channel) MarshalJSON() ([]byte, error) {
	type alias Channel
	if ch.Sampler == -1 {
		return json.Marshal(&struct {
			Sampler int32 `json:"sampler,omitempty"`
			*alias
		}{
			Sampler: 0,
			alias:   (*alias)(ch),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(ch)})
}

// ChannelTarget describes the index of the node and TRS property that an animation channel targets.
// The Path represents the name of the node's TRS property to modify, or the "weights" of the Morph Targets it instantiates.
// For the "translation" property, the values that are provided by the sampler are the translation along the x, y, and z axes.
// For the "rotation" property, the values are a quaternion in the order (x, y, z, w), where w is the scalar.
// For the "scale" property, the values are the scaling factors along the x, y, and z axes.
type ChannelTarget struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Node       int32       `json:"node" validate:"gte=-1"`
	Path       TRSProperty `json:"path" validate:"oneof=translation rotation scale weights"`
}

// NewChannelTarget returns a default ChannelTarget.
func NewChannelTarget(path TRSProperty) *ChannelTarget {
	return &ChannelTarget{Node: -1, Path: path}
}

// UnmarshalJSON unmarshal the channel target with the correct default values.
func (ch *ChannelTarget) UnmarshalJSON(data []byte) error {
	type alias ChannelTarget
	tmp := alias(*NewChannelTarget(""))
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*ch = ChannelTarget(tmp)
	}
	return err
}

// MarshalJSON marshal the channel target with the correct default values.
func (ch *ChannelTarget) MarshalJSON() ([]byte, error) {
	type alias ChannelTarget
	if ch.Node == -1 {
		return json.Marshal(&struct {
			Node int32 `json:"node,omitempty"`
			*alias
		}{
			Node:  0,
			alias: (*alias)(ch),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(ch)})
}

func removeProperty(str []byte, b []byte) []byte {
	b = bytes.Replace(b, str, []byte(""), 1)
	return bytes.Replace(b, []byte(`,,`), []byte(","), 1)
}

func sanitizeJSON(b []byte) []byte {
	b = bytes.Replace(b, []byte(`{,`), []byte("{"), 1)
	return bytes.Replace(b, []byte(`,}`), []byte("}"), 1)
}
