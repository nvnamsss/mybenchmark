package mybenchmark

import (
	"encoding/json"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"

	"mybenchmark/proto/serialpb"
)

// ===== Nested Struct Definitions =====

// Vec2 is a 2D vector with dual tags for JSON and msgpack serialization.
type Vec2 struct {
	X float64 `json:"x" msgpack:"x"`
	Y float64 `json:"y" msgpack:"y"`
}

// Vec3F is a 3D vector with dual tags for JSON and msgpack serialization.
type Vec3F struct {
	X float64 `json:"x" msgpack:"x"`
	Y float64 `json:"y" msgpack:"y"`
	Z float64 `json:"z" msgpack:"z"`
}

// PlayerStats groups character stats with dual tags.
type PlayerStats struct {
	Health float32 `json:"health" msgpack:"health"`
	Mana   float32 `json:"mana" msgpack:"mana"`
	Level  int32   `json:"level" msgpack:"level"`
	XP     int64   `json:"xp" msgpack:"xp"`
}

// Small nested struct: 5 fields organized into nested groups.
type UnitSkillEventNested struct {
	CasterID  int64 `json:"caster_id" msgpack:"caster_id"`
	AbilityID int64 `json:"ability_id" msgpack:"ability_id"`
	TargetID  int64 `json:"target_id" msgpack:"target_id"`
	Location  Vec2  `json:"location" msgpack:"location"`
}

// Large nested struct: 11 fields organized into sub-structs.
type PlayerStateNested struct {
	PlayerID int64       `json:"player_id" msgpack:"player_id"`
	Position Vec3F       `json:"position" msgpack:"position"`
	Velocity Vec3F       `json:"velocity" msgpack:"velocity"`
	Stats    PlayerStats `json:"stats" msgpack:"stats"`
}

// Test fixtures: pre-built sample structs.
var (
	unitSkillEventNestedSample = UnitSkillEventNested{
		CasterID:  1001,
		AbilityID: 2001,
		TargetID:  3001,
		Location: Vec2{
			X: 42.5,
			Y: 99.75,
		},
	}
	playerStateNestedSample = PlayerStateNested{
		PlayerID: 5001,
		Position: Vec3F{
			X: 10.5,
			Y: 20.25,
			Z: 30.75,
		},
		Velocity: Vec3F{
			X: 1.0,
			Y: 2.0,
			Z: 3.0,
		},
		Stats: PlayerStats{
			Health: 100.0,
			Mana:   50.5,
			Level:  42,
			XP:     1000000,
		},
	}

	// Protobuf fixtures: pointer-based proto messages with identical values.
	unitSkillEventNestedPBSample = &serialpb.UnitSkillEventNestedPB{
		CasterId:  1001,
		AbilityId: 2001,
		TargetId:  3001,
		Location: &serialpb.Vec2PB{
			X: 42.5,
			Y: 99.75,
		},
	}
	playerStateNestedPBSample = &serialpb.PlayerStateNestedPB{
		PlayerId: 5001,
		Position: &serialpb.Vec3FPB{
			X: 10.5,
			Y: 20.25,
			Z: 30.75,
		},
		Velocity: &serialpb.Vec3FPB{
			X: 1.0,
			Y: 2.0,
			Z: 3.0,
		},
		Stats: &serialpb.PlayerStatsPB{
			Health: 100.0,
			Mana:   50.5,
			Level:  42,
			Xp:     1000000,
		},
	}
)

// ===== Small Nested Struct - JSON =====

// BenchmarkNestedSmallJSONMarshal encodes a nested 5-field struct to JSON bytes.
func BenchmarkNestedSmallJSONMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = json.Marshal(unitSkillEventNestedSample)
	}
	_ = out
}

// BenchmarkNestedSmallJSONUnmarshal decodes JSON bytes into a nested 5-field struct.
func BenchmarkNestedSmallJSONUnmarshal(b *testing.B) {
	b.ReportAllocs()
	jsonBytes, _ := json.Marshal(unitSkillEventNestedSample)
	b.ResetTimer()
	for b.Loop() {
		var dst UnitSkillEventNested
		json.Unmarshal(jsonBytes, &dst)
	}
}

// ===== Small Nested Struct - msgpack =====

// BenchmarkNestedSmallMsgpackMarshal encodes a nested 5-field struct to msgpack bytes.
func BenchmarkNestedSmallMsgpackMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = msgpack.Marshal(unitSkillEventNestedSample)
	}
	_ = out
}

// BenchmarkNestedSmallMsgpackUnmarshal decodes msgpack bytes into a nested 5-field struct.
func BenchmarkNestedSmallMsgpackUnmarshal(b *testing.B) {
	b.ReportAllocs()
	msgpackBytes, _ := msgpack.Marshal(unitSkillEventNestedSample)
	b.ResetTimer()
	for b.Loop() {
		var dst UnitSkillEventNested
		msgpack.Unmarshal(msgpackBytes, &dst)
	}
}

// ===== Large Nested Struct - JSON =====

// BenchmarkNestedLargeJSONMarshal encodes a nested 11-field struct to JSON bytes.
func BenchmarkNestedLargeJSONMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = json.Marshal(playerStateNestedSample)
	}
	_ = out
}

// BenchmarkNestedLargeJSONUnmarshal decodes JSON bytes into a nested 11-field struct.
func BenchmarkNestedLargeJSONUnmarshal(b *testing.B) {
	b.ReportAllocs()
	jsonBytes, _ := json.Marshal(playerStateNestedSample)
	b.ResetTimer()
	for b.Loop() {
		var dst PlayerStateNested
		json.Unmarshal(jsonBytes, &dst)
	}
}

// ===== Large Nested Struct - msgpack =====

// BenchmarkNestedLargeMsgpackMarshal encodes a nested 11-field struct to msgpack bytes.
func BenchmarkNestedLargeMsgpackMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = msgpack.Marshal(playerStateNestedSample)
	}
	_ = out
}

// BenchmarkNestedLargeMsgpackUnmarshal decodes msgpack bytes into a nested 11-field struct.
func BenchmarkNestedLargeMsgpackUnmarshal(b *testing.B) {
	b.ReportAllocs()
	msgpackBytes, _ := msgpack.Marshal(playerStateNestedSample)
	b.ResetTimer()
	for b.Loop() {
		var dst PlayerStateNested
		msgpack.Unmarshal(msgpackBytes, &dst)
	}
}

// ===== Small Nested Struct - protobuf =====

// BenchmarkNestedSmallProtobufMarshal encodes a nested 5-field proto message to wire bytes.
func BenchmarkNestedSmallProtobufMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = proto.Marshal(unitSkillEventNestedPBSample)
	}
	_ = out
}

// BenchmarkNestedSmallProtobufUnmarshal decodes wire bytes into a nested 5-field proto message.
func BenchmarkNestedSmallProtobufUnmarshal(b *testing.B) {
	b.ReportAllocs()
	pbBytes, _ := proto.Marshal(unitSkillEventNestedPBSample)
	b.ResetTimer()
	for b.Loop() {
		var dst serialpb.UnitSkillEventNestedPB
		proto.Unmarshal(pbBytes, &dst)
	}
}

// ===== Large Nested Struct - protobuf =====

// BenchmarkNestedLargeProtobufMarshal encodes a nested 11-field proto message to wire bytes.
func BenchmarkNestedLargeProtobufMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = proto.Marshal(playerStateNestedPBSample)
	}
	_ = out
}

// BenchmarkNestedLargeProtobufUnmarshal decodes wire bytes into a nested 11-field proto message.
func BenchmarkNestedLargeProtobufUnmarshal(b *testing.B) {
	b.ReportAllocs()
	pbBytes, _ := proto.Marshal(playerStateNestedPBSample)
	b.ResetTimer()
	for b.Loop() {
		var dst serialpb.PlayerStateNestedPB
		proto.Unmarshal(pbBytes, &dst)
	}
}
