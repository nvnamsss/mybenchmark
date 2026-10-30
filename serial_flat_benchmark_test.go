package mybenchmark

import (
	"encoding/json"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"

	"mybenchmark/proto/serialpb"
)

// Small flat struct: 5 primitive fields (mirrors user's UnitSkillEvent example).
type UnitSkillEventFlat struct {
	CasterID  int64   `json:"caster_id" msgpack:"caster_id"`
	AbilityID int64   `json:"ability_id" msgpack:"ability_id"`
	TargetID  int64   `json:"target_id" msgpack:"target_id"`
	TargetX   float64 `json:"target_x" msgpack:"target_x"`
	TargetY   float64 `json:"target_y" msgpack:"target_y"`
}

// Large flat struct: 11 primitive fields representing a player state.
type PlayerStateFlat struct {
	PlayerID int64   `json:"player_id" msgpack:"player_id"`
	PosX     float64 `json:"pos_x" msgpack:"pos_x"`
	PosY     float64 `json:"pos_y" msgpack:"pos_y"`
	PosZ     float64 `json:"pos_z" msgpack:"pos_z"`
	VelX     float64 `json:"vel_x" msgpack:"vel_x"`
	VelY     float64 `json:"vel_y" msgpack:"vel_y"`
	VelZ     float64 `json:"vel_z" msgpack:"vel_z"`
	Health   float32 `json:"health" msgpack:"health"`
	Mana     float32 `json:"mana" msgpack:"mana"`
	Level    int32   `json:"level" msgpack:"level"`
	XP       int64   `json:"xp" msgpack:"xp"`
}

// Test fixtures: pre-built sample structs.
var (
	unitSkillEventFlatSample = UnitSkillEventFlat{
		CasterID:  1001,
		AbilityID: 2001,
		TargetID:  3001,
		TargetX:   42.5,
		TargetY:   99.75,
	}
	playerStateFlatSample = PlayerStateFlat{
		PlayerID: 5001,
		PosX:     10.5,
		PosY:     20.25,
		PosZ:     30.75,
		VelX:     1.0,
		VelY:     2.0,
		VelZ:     3.0,
		Health:   100.0,
		Mana:     50.5,
		Level:    42,
		XP:       1000000,
	}

	// Protobuf fixtures: pointer-based proto messages with identical values.
	unitSkillEventFlatPBSample = &serialpb.UnitSkillEventFlatPB{
		CasterId:  1001,
		AbilityId: 2001,
		TargetId:  3001,
		TargetX:   42.5,
		TargetY:   99.75,
	}
	playerStateFlatPBSample = &serialpb.PlayerStateFlatPB{
		PlayerId: 5001,
		PosX:     10.5,
		PosY:     20.25,
		PosZ:     30.75,
		VelX:     1.0,
		VelY:     2.0,
		VelZ:     3.0,
		Health:   100.0,
		Mana:     50.5,
		Level:    42,
		Xp:       1000000,
	}
)

// ===== Small Flat Struct - JSON =====

// BenchmarkFlatSmallJSONMarshal encodes a flat 5-field struct to JSON bytes.
func BenchmarkFlatSmallJSONMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = json.Marshal(unitSkillEventFlatSample)
	}
	_ = out
}

// BenchmarkFlatSmallJSONUnmarshal decodes JSON bytes into a flat 5-field struct.
func BenchmarkFlatSmallJSONUnmarshal(b *testing.B) {
	b.ReportAllocs()
	jsonBytes, _ := json.Marshal(unitSkillEventFlatSample)
	b.ResetTimer()
	for b.Loop() {
		var dst UnitSkillEventFlat
		json.Unmarshal(jsonBytes, &dst)
	}
}

// ===== Small Flat Struct - msgpack =====

// BenchmarkFlatSmallMsgpackMarshal encodes a flat 5-field struct to msgpack bytes.
func BenchmarkFlatSmallMsgpackMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = msgpack.Marshal(unitSkillEventFlatSample)
	}
	_ = out
}

// BenchmarkFlatSmallMsgpackUnmarshal decodes msgpack bytes into a flat 5-field struct.
func BenchmarkFlatSmallMsgpackUnmarshal(b *testing.B) {
	b.ReportAllocs()
	msgpackBytes, _ := msgpack.Marshal(unitSkillEventFlatSample)
	b.ResetTimer()
	for b.Loop() {
		var dst UnitSkillEventFlat
		msgpack.Unmarshal(msgpackBytes, &dst)
	}
}

// ===== Large Flat Struct - JSON =====

// BenchmarkFlatLargeJSONMarshal encodes a flat 11-field struct to JSON bytes.
func BenchmarkFlatLargeJSONMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = json.Marshal(playerStateFlatSample)
	}
	_ = out
}

// BenchmarkFlatLargeJSONUnmarshal decodes JSON bytes into a flat 11-field struct.
func BenchmarkFlatLargeJSONUnmarshal(b *testing.B) {
	b.ReportAllocs()
	jsonBytes, _ := json.Marshal(playerStateFlatSample)
	b.ResetTimer()
	for b.Loop() {
		var dst PlayerStateFlat
		json.Unmarshal(jsonBytes, &dst)
	}
}

// ===== Large Flat Struct - msgpack =====

// BenchmarkFlatLargeMsgpackMarshal encodes a flat 11-field struct to msgpack bytes.
func BenchmarkFlatLargeMsgpackMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = msgpack.Marshal(playerStateFlatSample)
	}
	_ = out
}

// BenchmarkFlatLargeMsgpackUnmarshal decodes msgpack bytes into a flat 11-field struct.
func BenchmarkFlatLargeMsgpackUnmarshal(b *testing.B) {
	b.ReportAllocs()
	msgpackBytes, _ := msgpack.Marshal(playerStateFlatSample)
	b.ResetTimer()
	for b.Loop() {
		var dst PlayerStateFlat
		msgpack.Unmarshal(msgpackBytes, &dst)
	}
}

// ===== Small Flat Struct - protobuf =====

// BenchmarkFlatSmallProtobufMarshal encodes a flat 5-field proto message to wire bytes.
func BenchmarkFlatSmallProtobufMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = proto.Marshal(unitSkillEventFlatPBSample)
	}
	_ = out
}

// BenchmarkFlatSmallProtobufUnmarshal decodes wire bytes into a flat 5-field proto message.
func BenchmarkFlatSmallProtobufUnmarshal(b *testing.B) {
	b.ReportAllocs()
	pbBytes, _ := proto.Marshal(unitSkillEventFlatPBSample)
	b.ResetTimer()
	for b.Loop() {
		var dst serialpb.UnitSkillEventFlatPB
		proto.Unmarshal(pbBytes, &dst)
	}
}

// ===== Large Flat Struct - protobuf =====

// BenchmarkFlatLargeProtobufMarshal encodes a flat 11-field proto message to wire bytes.
func BenchmarkFlatLargeProtobufMarshal(b *testing.B) {
	b.ReportAllocs()
	var out []byte
	for b.Loop() {
		out, _ = proto.Marshal(playerStateFlatPBSample)
	}
	_ = out
}

// BenchmarkFlatLargeProtobufUnmarshal decodes wire bytes into a flat 11-field proto message.
func BenchmarkFlatLargeProtobufUnmarshal(b *testing.B) {
	b.ReportAllocs()
	pbBytes, _ := proto.Marshal(playerStateFlatPBSample)
	b.ResetTimer()
	for b.Loop() {
		var dst serialpb.PlayerStateFlatPB
		proto.Unmarshal(pbBytes, &dst)
	}
}
