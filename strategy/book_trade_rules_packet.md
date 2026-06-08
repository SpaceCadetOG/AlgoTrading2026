# Book Trade Rules Packet

Playbooks: 9

Execution Enabled: false

Paper Trading Enabled: false

Status: ready_for_paper_engine

## VP Reversal + OF Absorption

Setup:
Failed auction or sharp rejection at VAH, VAL, HVN, or POC with profile structure showing reversal context.

Entry:
Enter after rejection confirms and absorption, big limit order behavior, or cumulative delta divergence validates the reversal.

Stop:
Beyond the failed auction extreme or rejection wick, using the farthest meaningful barrier between profile edge and swing.

Target:
TP1: POC
TP2: Opposite value area boundary
TP3: Next HVN with trailing runner

Management:
Break-even after: TP1 or first clean reaction away from the failed level
Trail after: TP2 once reversal expands beyond value
What confirms it:
- Absorption or big limit order
- Cumulative delta divergence or failed continuation read

What cancels it:
- Price accepts beyond the failed level
- Order flow confirms the original move instead of the reversal

Risk:
Risk per trade: 0.50%
Min reward-to-risk: 1.25

## VP Accumulation + OF Volume Cluster

Setup:
Sideways accumulation zone forms with a POC or HVN inside the range and a breakout/retest context around a volume cluster.

Entry:
Enter after breakout from accumulation and retest of POC, HVN, or cluster confirms hold from the new side.

Stop:
Beyond the opposite side of the accumulation range or behind VAL for longs and VAH for shorts, with a small basis-point buffer.

Target:
TP1: Nearest HVN or first profile barrier
TP2: Next major profile level
TP3: Trend runner if expansion persists

Management:
Break-even after: TP1
Trail after: New HVNs form in trend direction
What confirms it:
- Breakout plus retest hold
- Volume cluster acting as support or resistance

What cancels it:
- Price accepts back into the accumulation range
- Retest slices through the level without reaction

Risk:
Risk per trade: 0.50%
Min reward-to-risk: 1.25

## VP Trend + VWAP Pullback + Aggressive Delta

Setup:
Trend profile is directional, VWAP is aligned with the move, and pullback reaches a profile level, VWAP, or prior acceptance zone.

Entry:
Enter after pullback to VWAP, deviation, or profile level holds and aggressive delta resumes in the trend direction.

Stop:
Beyond the pullback swing, VWAP barrier, or profile level that should hold, whichever is farthest and structurally cleanest.

Target:
TP1: Fixed 1R
TP2: Next HVN or VWAP deviation
TP3: Trailing trend target

Management:
Break-even after: TP1
Trail after: TP1 or TP2 when continuation proves itself
What confirms it:
- Aggressive delta with trend
- Continuation candle or reaction confirmation

What cancels it:
- Price accepts back into prior value
- Delta flips against continuation at the retest

Risk:
Risk per trade: 0.50%
Min reward-to-risk: 1.25

## VP Rejection + Cumulative Delta Divergence

Setup:
Price tests VAH, VAL, HVN, LVN, or POC and rejects, while cumulative delta diverges from price at the rejection point.

Entry:
Enter after the rejection candle confirms and cumulative delta divergence shows failed participation in the tested direction.

Stop:
Beyond the rejection wick or failed auction extreme, with a buffer beyond the level that should hold.

Target:
TP1: POC
TP2: Opposite value area boundary
TP3: Next HVN beyond value

Management:
Break-even after: TP1
Trail after: Only after TP2 if rotation expands
What confirms it:
- Cumulative delta divergence
- Rejection candle or retest fail

What cancels it:
- Price accepts past the rejected level
- Divergence disappears on the next push

Risk:
Risk per trade: 0.50%
Min reward-to-risk: 1.25

## Unfinished Business Revisit

Setup:
An unfinished high or low exists and price later revisits that level, creating a chance to trade the revisit after confirmation rather than the first print.

Entry:
Enter only after the unfinished level is revisited and held or rejected with clear response.

Stop:
Beyond the unfinished level with a small buffer beyond the auction extreme.

Target:
TP1: Nearest HVN or POC
TP2: Auction completion objective
TP3: VWAP or next meaningful profile barrier

Management:
Break-even after: First meaningful reaction away from the unfinished level
Trail after: Only after the auction completion move is underway
What confirms it:
- Revisit reaction

What cancels it:
- Revisit chops through the level without clear response
- Liquidity disappears and the revisit loses structure

Risk:
Risk per trade: 0.50%
Min reward-to-risk: 1.25

## VWAP First Touch Combo

Setup:
Price approaches VWAP from the correct side and the first touch lines up with a volume-profile level, order-flow confirmation, or price-action structure.

Entry:
Enter on the first VWAP touch only when confluence exists with profile or order-flow context.

Stop:
Beyond VWAP plus the recent swing or ATR-style structural buffer, choosing the farthest meaningful barrier.

Target:
TP1: 1R starter target
TP2: Next VWAP deviation or profile barrier
TP3: Optional runner into broader structure

Management:
Break-even after: Clear favorable reaction or TP1
Trail after: Only after TP1 if move proves itself
What confirms it:
- Confluence confirmation at touch

What cancels it:
- No confluence at VWAP
- VWAP is touched in chop without reaction

Risk:
Risk per trade: 0.50%
Min reward-to-risk: 1.25

## VWAP Successful Reaction

Setup:
Price touches VWAP and reacts away, creating a safer entry than a blind first touch.

Entry:
Enter after the VWAP touch reacts and a confirmation candle closes in the reaction direction.

Stop:
Beyond the reaction swing or just beyond VWAP if that is the level that should hold.

Target:
TP1: 1R
TP2: VWAP first deviation
TP3: Next volume-profile barrier

Management:
Break-even after: Clear favorable follow-through or TP1
Trail after: After TP1 only
What confirms it:
- Price confirmation candle

What cancels it:
- Price accepts through VWAP after touch

Risk:
Risk per trade: 0.50%
Min reward-to-risk: 1.25

## VWAP Deviation Rotation

Setup:
Price rotates between VWAP and the first deviation in a range, creating fade opportunities back toward fair value.

Entry:
Fade the first deviation back toward VWAP after price reacts from the deviation instead of accepting beyond it.

Stop:
Beyond the deviation extreme that should not be accepted if rotation remains valid.

Target:
TP1: VWAP midpoint or line
TP2: Opposite first deviation
TP3: Optional full range extension if rotation persists

Management:
Break-even after: Return toward VWAP
Trail after: Only after VWAP is reclaimed and move continues
What confirms it:
- Reaction from first deviation

What cancels it:
- Price accepts beyond the deviation
- Trend conditions replace rotational conditions

Risk:
Risk per trade: 0.50%
Min reward-to-risk: 1.25

## VWAP Trend Deviation Continuation

Setup:
Strong trend pulls back to the first VWAP deviation and uses that deviation as continuation support or resistance.

Entry:
Enter after pullback to first deviation confirms continuation with supportive price action or aligned order-flow behavior.

Stop:
Beyond the deviation and the pullback swing, using the strongest barrier among deviation, swing, profile level, and ATR-style fallback.

Target:
TP1: 1R
TP2: Next deviation or extension target
TP3: Trailing trend target

Management:
Break-even after: TP1 or immediate favorable continuation
Trail after: TP2 or once the extension is underway
What confirms it:
- Continuation confirmation

What cancels it:
- Price accepts beyond the deviation
- Momentum fades or delta diverges against the move

Risk:
Risk per trade: 0.50%
Min reward-to-risk: 1.25

