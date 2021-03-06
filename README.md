# gltf
[![Documentation](https://godoc.org/github.com/qmuntal/gltf?status.svg)](https://godoc.org/github.com/qmuntal/gltf)
[![Build Status](https://travis-ci.com/qmuntal/gltf.svg?branch=master)](https://travis-ci.com/qmuntal/gltf)
[![Go Report Card](https://goreportcard.com/badge/github.com/qmuntal/gltf)](https://goreportcard.com/report/github.com/qmuntal/gltf)
[![codecov](https://coveralls.io/repos/github/qmuntal/gltf/badge.svg)](https://coveralls.io/github/qmuntal/gltf?branch=master)
[![codeclimate](https://codeclimate.com/github/qmuntal/gltf/badges/gpa.svg)](https://codeclimate.com/github/qmuntal/gltf)
[![License](https://img.shields.io/badge/License-BSD%202--Clause-orange.svg)](https://opensource.org/licenses/BSD-2-Clause)

A Go package for simple, efficient, and robust serialization/deserialization of [glTF 2.0](https://www.khronos.org/gltf/) (GL Transmission Format), a royalty-free specification for the efficient transmission and loading of 3D scenes and models by applications.

## Features
* High parsing time and moderate memory consumption.
* glTF specification v2.0.0
  * [x] ASCII glTF.
  * [x] Binary glTF(GLB).
  * [x] PBR material description.
* glTF validaton
  * [x] Validate against schemas.
  * [ ] Validate coherence.
* Buffers
  * [x] Parse BASE64 encoded embedded buffer data(DataURI).
  * [x] Load .bin file.
* Read from io.Reader
  * [x] Boilerplate for disk loading.
  * [x] Custom callback handlers.
  * [x] Automatic ASCII / glTF detection.
* Write to io.Writer
  * [x] Boilerplate for disk saving.
  * [x] Custom callback handlers.
  * [x] ASCII / Binary
* Extensions
  * [ ] KHR_draco_mesh_compression
  * [ ] KHR_lights_punctual
  * [x] KHR_materials_pbrSpecularGlossiness
  * [ ] KHR_materials_unlit
  * [ ] KHR_techniques_webgl
  * [ ] KHR_texture_transform

## Perfomance
All the functionality is benchmarked and tested using the official [glTF Samples](https://github.com/KhronosGroup/glTF-Sample-Models) in the utility package [qmuntal/gltf-bench](https://github.com/qmuntal/gltf-bench/).
The results show that the perfomance of this package is equivalent to [fx-gltf](https://github.com/jessey-git/fx-gltf), a reference perfomance-driven glTF implementation for C++, .

## Examples
### Read
```go
doc, err := gltf.Open("./a.gltf")
if err != nil {
  panic(err)
}
fmt.Print(doc.Asset)
```
### Write
```go
doc := &gltf.Document{
  Scene: 0, 
  Asset: gltf.Asset{Generator: "qmuntal/gltf"}, 
  Scenes: []gltf.Scene{{Extras: 8.0, Extensions: gltf.Extensions{"a": "b"}, Name: "s_1"}}
}
 
if err := gltf.Save(doc, "./a.gltf", true); err != nil {
  panic(err)
}
```

### Write complex
```go
doc := &gltf.Document{
  Accessors: []gltf.Accessor{
    {BufferView: 0, ByteOffset: 0, ComponentType: gltf.UnsignedShort, Count: 36, Type: gltf.Scalar},
    {BufferView: 1, ByteOffset: 0, ComponentType: gltf.Float, Count: 24, Max: []float64{0.5, 0.5, 0.5}, Min: []float64{-0.5, -0.5, -0.5}, Type: gltf.Vec3},
    {BufferView: 2, ByteOffset: 0, ComponentType: gltf.Float, Count: 24, Type: gltf.Vec3},
    {BufferView: 3, ByteOffset: 0, ComponentType: gltf.Float, Count: 24, Type: gltf.Vec4},
    {BufferView: 4, ByteOffset: 0, ComponentType: gltf.Float, Count: 24, Type: gltf.Vec2},
  },
  Asset: gltf.Asset{Version: "2.0", Generator: "FBX2glTF"},
  BufferViews: []gltf.BufferView{
    {Buffer: 0, ByteLength: 72, ByteOffset: 0, Target: gltf.ElementArrayBuffer},
    {Buffer: 0, ByteLength: 288, ByteOffset: 72, Target: gltf.ArrayBuffer},
    {Buffer: 0, ByteLength: 288, ByteOffset: 360, Target: gltf.ArrayBuffer},
    {Buffer: 0, ByteLength: 384, ByteOffset: 648, Target: gltf.ArrayBuffer},
    {Buffer: 0, ByteLength: 192, ByteOffset: 1032, Target: gltf.ArrayBuffer},
  },
  Buffers: []gltf.Buffer{{ByteLength: 1224, Data: readFile("testdata/BoxVertexColors/glTF-Binary/BoxVertexColors.glb")[1628+20+8:]}},
  Materials: []gltf.Material{{Name: "Default", AlphaMode: gltf.Opaque, AlphaCutoff: 0.5, 
    PBRMetallicRoughness: &gltf.PBRMetallicRoughness{BaseColorFactor: [4]float64{0.8, 0.8, 0.8, 1}, MetallicFactor: 0.1, RoughnessFactor: 0.99}}
  },
  Meshes: []gltf.Mesh{{Name: "Cube", Primitives: []gltf.Primitive{{Indices: 0, Material: 0, Mode: gltf.Triangles, Attributes: map[string]uint32{"POSITION": 1, "COLOR_0": 3, "NORMAL": 2, "TEXCOORD_0": 4}}}}},
  Nodes: []gltf.Node{
    {Name: "RootNode", Mesh: -1, Camera: -1, Skin: -1, Children: []uint32{1, 2, 3}, Matrix: [16]float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}, Rotation: [4]float64{0, 0, 0, 1}, Scale: [3]float64{1, 1, 1}},
    {Name: "Mesh", Mesh: -1, Camera: -1, Skin: -1, Matrix: [16]float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}, Rotation: [4]float64{0, 0, 0, 1}, Scale: [3]float64{1, 1, 1}},
    {Name: "Cube", Mesh: 0, Camera: -1, Skin: -1, Matrix: [16]float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}, Rotation: [4]float64{0, 0, 0, 1}, Scale: [3]float64{1, 1, 1}},
    {Name: "Texture Group", Mesh: -1, Camera: -1, Skin: -1, Matrix: [16]float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}, Rotation: [4]float64{0, 0, 0, 1}, Scale: [3]float64{1, 1, 1}},
  },
  Samplers: []gltf.Sampler{{WrapS: gltf.Repeat, WrapT: gltf.Repeat}},
  Scene: 0,
  Scenes: []gltf.Scene{{Name: "Root Scene", Nodes: []uint32{0}}},
}
if err := gltf.Save(doc, "./a.gltf", true); err != nil {
  panic(err)
}
```
