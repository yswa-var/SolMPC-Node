package businesslogic

import "fmt"

type BusinessRules struct {
	ProtocolFeeBps uint16 // e.g., 200 (2%)
	CuratorBps     uint16 // e.g., 2000 (20%)
	PublisherBps   uint16 // e.g., 1000 (10%)
	SubTiltRules   []SubTiltRule
}

type SubTiltRule struct {
	TiltID   string
	ShareBps uint16 // Percentage of the curator's share
}

type Distribution struct {
	Recipients []string
	Amounts    []uint64
}

func computeDistribution(balance uint64, rules BusinessRules) Distribution {
	protocolShare := balance * uint64(rules.ProtocolFeeBps) / 10000
	curatorShare := balance * uint64(rules.CuratorBps) / 10000
	publisherShare := balance * uint64(rules.PublisherBps) / 10000
	remainingBalance := balance - (protocolShare + curatorShare + publisherShare)

	recipients := []string{"Protocol", "Publisher"}
	amounts := []uint64{protocolShare, publisherShare}

	// Split curator share among sub-tilts
	for _, subTilt := range rules.SubTiltRules {
		subAmount := curatorShare * uint64(subTilt.ShareBps) / 10000
		recipients = append(recipients, subTilt.TiltID)
		amounts = append(amounts, subAmount)
	}

	// If any remaining balance, assign it to main curator
	if remainingBalance > 0 {
		recipients = append(recipients, "Main Curator")
		amounts = append(amounts, remainingBalance)
	}

	return Distribution{Recipients: recipients, Amounts: amounts}
}

func test() {
	// Example dummy rules
	dummyRules := BusinessRules{
		ProtocolFeeBps: 200,  // 2%
		CuratorBps:     2000, // 20%
		PublisherBps:   1000, // 10%
		SubTiltRules: []SubTiltRule{
			{TiltID: "SubTiltA", ShareBps: 5000}, // 50% of curator's share
			{TiltID: "SubTiltB", ShareBps: 5000}, // 50% of curator's share
		},
	}

	totalBalance := uint64(1000000) // Example: 1,000,000 lamports
	result := computeDistribution(totalBalance, dummyRules)
	for i, recipient := range result.Recipients {
		fmt.Printf("%s: %d lamports\n", recipient, result.Amounts[i])
	}
}
