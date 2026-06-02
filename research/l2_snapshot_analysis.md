# Historical L2 Snapshot Analysis

## Dataset Quality

- totalRows: 72
- validRows: 72
- invalidRows: 0
- errorRows: 0
- validPct: 100.00
- errorPct: 0.00

## Venue Summaries

| Venue | Snapshots | Valid | Invalid | Avg Spread % | Min Spread % | Max Spread % | Avg Imbalance | Min Imbalance | Max Imbalance | Avg Bid Depth | Avg Ask Depth | Avg Mid |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| aster | 24 | 24 | 0 | 0.00033622 | 0.00013663 | 0.00328067 | -0.00203488 | -0.17273561 | 0.20477187 | 6640097.75 | 6677398.58 | 73116.64 |
| hyperliquid | 24 | 24 | 0 | 0.00136766 | 0.00136643 | 0.00136985 | 0.08028955 | -0.53274204 | 0.91563225 | 6824121.52 | 5554252.61 | 73117.83 |
| lighter | 24 | 24 | 0 | 0.00234250 | 0.00013666 | 0.00931247 | 0.03687703 | -0.12203331 | 0.13058190 | 44794494.52 | 41621943.18 | 73110.48 |

## Cross Venue Comparison

- comparableGroups: 24
- averageMidDifference: 12.56250000
- maxMidDifference: 24.35000000
- averageSpreadDifference: 0.00286663
- widestSpreadVenue: lighter
- tightestSpreadVenue: aster
- deepestLiquidityVenue: lighter

## Liquidity Rankings

### Average Bid Depth

| Rank | Venue | Value |
|---:|---|---:|
| 1 | lighter | 44794494.51722492 |
| 2 | hyperliquid | 6824121.51995875 |
| 3 | aster | 6640097.74794167 |

### Average Ask Depth

| Rank | Venue | Value |
|---:|---|---:|
| 1 | lighter | 41621943.18227631 |
| 2 | aster | 6677398.58390000 |
| 3 | hyperliquid | 5554252.61071458 |

### Combined Depth

| Rank | Venue | Value |
|---:|---|---:|
| 1 | lighter | 86416437.69950122 |
| 2 | aster | 13317496.33184167 |
| 3 | hyperliquid | 12378374.13067333 |

### Tightest Spreads

| Rank | Venue | Value |
|---:|---|---:|
| 1 | aster | 0.00033622 |
| 2 | hyperliquid | 0.00136766 |
| 3 | lighter | 0.00234250 |

### Highest Bid Pressure

| Rank | Venue | Value |
|---:|---|---:|
| 1 | hyperliquid | 0.08028955 |
| 2 | lighter | 0.03687703 |
| 3 | aster | -0.00203488 |

### Highest Ask Pressure

| Rank | Venue | Value |
|---:|---|---:|
| 1 | aster | -0.00203488 |
| 2 | lighter | 0.03687703 |
| 3 | hyperliquid | 0.08028955 |


## Findings

- Tightest spread venue: aster.
- Widest spread venue: lighter.
- Deepest liquidity venue: lighter.
- Highest bid pressure venue: hyperliquid (avg imbalance 0.0803).
- Highest ask pressure venue: aster (avg imbalance -0.0020).
- Recorder output looks stable enough for a longer pilot collection.

## Notes

- Historical L2 recorder Phase 1 samples snapshots at fixed intervals; it is not tick-by-tick book replay.
- Cross-venue comparison groups rows by recorder round order because venue timestamps are collected sequentially, not simultaneously.
- CSV storage is local and append-only; larger recordings should manage file rotation and metadata.
