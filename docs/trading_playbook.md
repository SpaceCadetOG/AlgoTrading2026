# Trading Playbook

These playbooks map the book research into structured bot-ready rules without enabling execution, OMS wiring, paper trading, or live trading.

## VP Reversal + Order Flow Absorption

Setup:
A reversal profile forms around rejection or failed continuation, with price rejecting value near POC, VAH, VAL, HVN, LVN, or a prior high/low.

Entry:
Enter only after absorption confirms that aggressive flow failed to continue and price reclaims the reversal level in the new direction.
Entry Zone: Reversal profile POC first, then nearest HVN or VAH/VAL boundary if POC is already reclaimed.

Stop:
Beyond the rejection extreme, with the value-area boundary as the structural line in the sand.
What cancels it: Long idea is invalid if price closes back below the rejection low or below VAL. Short idea is invalid if price closes back above the rejection high or above VAH.

Target:
First target: Return to the opposite side of the local value area or the nearest opposing HVN.
Second target: Expansion toward the next higher-timeframe POC, VAH, or prior swing objective.

Management:
Exit early if price stalls back inside value after absorption confirmation or if follow-through fails across the first few bars.
- Reduce risk after the first clean move away from the reversal level.
- Treat this as a context trade first; confirmation must stay aligned after entry.

What confirms it:
- Absorption proxy aligned with the reversal direction
- Price response away from the rejected side

What cancels it:
- Price accepts back through the rejected value area.
- Absorption flips against the intended direction.

Required context:
- Volume profile reversal structure
- Failed auction or strong rejection context
- Nearby structural level such as POC, HVN, VAH, VAL, or prior high/low

## VP Accumulation + Volume Cluster

Setup:
A sideways accumulation profile forms with a valid POC or dominant HVN, then price leaves the range with aggressive initiation.

Entry:
Wait for price to return to the accumulation POC or dominant volume cluster and hold it as support for longs or resistance for shorts.
Entry Zone: Accumulation POC first, then the nearest qualifying HVN inside the accumulation range.

Stop:
Outside the accumulation value area, beyond the side that should no longer be revisited.
What cancels it: Long idea is invalid below VAL or below the accumulation shelf. Short idea is invalid above VAH or above the accumulation ceiling.

Target:
First target: The range expansion objective or the next visible HVN above/below the setup level.
Second target: The next broader POC confluence or the prior directional swing extension.

Management:
Exit if the retest fails to hold the cluster level, if acceptance degrades quickly, or if delta turns sharply against the breakout direction.
- Stronger confluence exists when accumulation POC aligns with daily, rolling 3-day, rolling 7-day, or composite POC.
- Use cluster retest quality as the filter before expanding to backtest mapping.

What confirms it:
- Retest of the accumulation level
- Volume cluster hold at the setup level

What cancels it:
- Retest slices through the accumulation level without hold.
- Profile shape loses acceptance and price rotates back through the range.

Required context:
- Sideways accumulation profile
- Aggressive initiation out of the range
- Volume cluster or accumulation POC

## VP Trend + Aggressive Delta

Setup:
A directional trend leg forms from aggressive initiation or open-drive behavior, with a trend-leg profile identifying the key POC or HVN.

Entry:
Enter on a pullback that retests the trend-leg POC or nearest HVN while aggressive delta reconfirms continuation.
Entry Zone: Trend-leg POC first, nearest HVN second, ideally with VWAP still aligned to trend.

Stop:
Just outside the trend-leg value area, beyond the level that should hold if the trend is intact.
What cancels it: Bull trend invalidates below VAL. Bear trend invalidates above VAH.

Target:
First target: Resume toward the most recent trend extreme.
Second target: Extension into the next trend objective, such as the next HVN gap fill or broader directional target.

Management:
Exit if aggressive delta fades immediately after entry or if the pullback turns into acceptance inside the trend profile.
- Trend strength score matters more than raw sample count for this playbook.
- Keep this mapped as continuation logic, not a blind momentum chase.

What confirms it:
- Aggressive delta in the trend direction
- Clean retest of the trend-leg level

What cancels it:
- Aggressive delta prints in the opposite direction at the retest.
- Price accepts through the trend-leg value area instead of continuing.

Required context:
- Directional trend leg profile
- Trend-leg POC or HVN
- VWAP aligned with trend if available

## VP Rejection + Cumulative Delta Divergence

Setup:
A rejection profile forms after higher or lower prices are sharply rejected, and cumulative delta diverges from price at the retest or exhaustion point.

Entry:
Enter when the rejection level holds and cumulative delta fails to confirm the attempted continuation.
Entry Zone: Rejection profile POC first, then nearest HVN or the rejected VAH/VAL edge.

Stop:
Beyond the rejection edge that should stay defended if the divergence is real.
What cancels it: Long idea is invalid below the rejection low or below VAL. Short idea is invalid above the rejection high or above VAH.

Target:
First target: Rotation back toward the profile POC or opposite side of local value.
Second target: Continuation into the next broader profile reference or prior swing.

Management:
Exit if divergence disappears on the next few bars or if price starts accepting beyond the rejection zone.
- Cumulative delta divergence is a confirmation layer, not a standalone trigger.
- Use profile context to avoid treating every divergence as tradable.

What confirms it:
- Cumulative delta divergence
- Hold at the rejection level

What cancels it:
- Cumulative delta reconfirms the original move instead of diverging.
- Price gains acceptance beyond the rejection boundary.

Required context:
- Rejection profile
- Strong rejection or failed continuation
- Nearby POC, HVN, VAH, or VAL

## Unfinished Business Revisit

Setup:
An unfinished-business high or low is marked, then price revisits that extreme after the initial magnet behavior is established.

Entry:
Enter only after the revisit confirms either rejection of the unfinished extreme or acceptance through it, depending on the side being tested.
Entry Zone: The unfinished-business extreme itself, with nearby footprint structure used to refine the revisit level.

Stop:
Just beyond the unfinished high or low that defines the revisit thesis.
What cancels it: Invalidate once price cleanly accepts through the unfinished-business level in the wrong direction.

Target:
First target: Return toward the nearest internal footprint HVN or recent range midpoint.
Second target: Continuation toward the opposite side of the intraday structure if the revisit resolves cleanly.

Management:
Exit if the revisit loses magnet behavior and stalls immediately after the test.
- Treat unfinished business as a revisit framework, not an instant reversal command.
- Pair with surrounding footprint and profile structure before mapping to automation.

What confirms it:
- Revisit of the unfinished level
- Clear hold or fail response after the revisit

What cancels it:
- The revisit never confirms hold or fail behavior.
- Price chops around the unfinished level without directional response.

Required context:
- Recorded unfinished-business high or low
- Nearby footprint structure

