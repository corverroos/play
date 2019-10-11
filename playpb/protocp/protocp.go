package protocp

import (
	"github.com/corverroos/play"
	pb "github.com/corverroos/play/playpb"
)

func RoundDataToProto(p *play.RoundData) *pb.RoundData {
	m := make(map[int64]int64)
	for i, part := range p.Parts {
		m[int64(i)] = int64(part)
	}
	return &pb.RoundData{
		ExternalID: p.ExternalID,
		Included:   p.Included,
		Submitted:  p.Submitted,
		Rank:       int64(p.Rank),
		Parts:      m,
	}
}

func RoundDataFromProto(p *pb.RoundData) play.RoundData {
	m := make(map[int]int)
	for i, part := range p.Parts {
		m[int(i)] = int(part)
	}
	return play.RoundData{
		ExternalID: p.ExternalID,
		Included:   p.Included,
		Submitted:  p.Submitted,
		Rank:       int(p.Rank),
		Parts:      m,
	}
}
