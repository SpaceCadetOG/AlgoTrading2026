package strategy

import "testing"

func TestDefaultBookTradeRules(t *testing.T) {
	playbooks := DefaultBookTradeRules()
	if len(playbooks) != 9 {
		t.Fatalf("expected 9 playbooks, got %d", len(playbooks))
	}
	for _, playbook := range playbooks {
		if playbook.Entry.Name == "" {
			t.Fatalf("playbook %q missing entry rule", playbook.Name)
		}
		if playbook.Stop.Name == "" {
			t.Fatalf("playbook %q missing stop rule", playbook.Name)
		}
		if playbook.Target.Name == "" {
			t.Fatalf("playbook %q missing target rule", playbook.Name)
		}
		if playbook.Management.Name == "" {
			t.Fatalf("playbook %q missing management rule", playbook.Name)
		}
		if playbook.Risk.Name == "" {
			t.Fatalf("playbook %q missing risk rule", playbook.Name)
		}
		if len(playbook.RequiredContext) == 0 {
			t.Fatalf("playbook %q missing required context", playbook.Name)
		}
		if len(playbook.RequiredConfirmations) == 0 {
			t.Fatalf("playbook %q missing required confirmations", playbook.Name)
		}
		if playbook.Target.MinRR < 1.0 {
			t.Fatalf("playbook %q min RR below 1.0: %.2f", playbook.Name, playbook.Target.MinRR)
		}
	}
}

func TestDefaultBookTradeRulesPacketSafetyFlags(t *testing.T) {
	packet := DefaultBookTradeRulesPacket()
	if packet.ExecutionEnabled {
		t.Fatal("execution should remain disabled")
	}
	if packet.PaperTradingEnabled {
		t.Fatal("paper trading should remain disabled")
	}
	if packet.Status != "ready_for_paper_engine" {
		t.Fatalf("unexpected status: %s", packet.Status)
	}
}
