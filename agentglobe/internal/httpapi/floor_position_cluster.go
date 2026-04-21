package httpapi

import (
	"strings"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
)

var floorInferredClusterTokens = map[string]struct{}{
	"long": {}, "short": {}, "neutral": {}, "speculative": {}, "unclustered": {},
}

// floorQueryClusterIsInferredStyle is true when `cluster` query values should filter inferred_cluster_at_stake
// rather than regional_cluster (e.g. CN-cluster stays regional).
func floorQueryClusterIsInferredStyle(cluster string) bool {
	_, ok := floorInferredClusterTokens[strings.ToLower(strings.TrimSpace(cluster))]
	return ok
}

func floorNormalizeInferredCluster(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "long":
		return "long"
	case "short":
		return "short"
	case "neutral", "abstain", "abstention":
		return "neutral"
	case "speculative", "spec":
		return "speculative"
	case "unclustered":
		return "unclustered"
	default:
		return ""
	}
}

// floorPositionLegacySpeculativeDirection is true when direction stored the old way (speculative as a "direction").
func floorPositionLegacySpeculativeDirection(dir string) bool {
	d := strings.ToLower(strings.TrimSpace(dir))
	return d == "speculative" || d == "spec"
}

// floorPositionBaseDirection returns long | short | neutral for the economic call; speculative is not a base direction.
func floorPositionBaseDirection(dir string) string {
	d := strings.ToLower(strings.TrimSpace(dir))
	switch d {
	case "long":
		return "long"
	case "short":
		return "short"
	case "speculative", "spec":
		return "neutral"
	case "neutral", "abstain", "abstention":
		return "neutral"
	default:
		return "neutral"
	}
}

func floorPositionSpeculativeFlag(p *dbpkg.FloorPosition) bool {
	if p == nil {
		return false
	}
	if p.Speculative {
		return true
	}
	return floorPositionLegacySpeculativeDirection(p.Direction)
}

// floorPositionInferredClusterForAggregate derives long|short|neutral|speculative|unclustered for rollups
// (discovery, cluster mix). Stored inferred_cluster_at_stake wins when set and valid.
func floorPositionInferredClusterForAggregate(p *dbpkg.FloorPosition) string {
	if p == nil {
		return "neutral"
	}
	if p.InferredClusterAtStake != nil {
		if c := floorNormalizeInferredCluster(*p.InferredClusterAtStake); c != "" {
			return c
		}
	}
	if floorPositionSpeculativeFlag(p) {
		return "speculative"
	}
	switch floorPositionBaseDirection(p.Direction) {
	case "long":
		return "long"
	case "short":
		return "short"
	default:
		return "neutral"
	}
}
