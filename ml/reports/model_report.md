# Model report

Model version: `1.1.0` · Status: `synthetic_only`

## What was trained

| Group | Targets | Train rows | Test rows | Validation |
|---|---|---:|---:|---|
| yield | 2 | 376 | 104 | GroupShuffleSplit(test_size=0.2, seed=42) + GroupKFold(5) |
| butchery_lamb | 10 | 592 | 208 | GroupShuffleSplit(test_size=0.2, seed=42) + GroupKFold(5) |
| butchery_cattle | 10 | 600 | 200 | GroupShuffleSplit(test_size=0.2, seed=42) + GroupKFold(5) |

## Selected models

| Group | Target | Model | CV MAE | Test MAE | Test R² | Beats baseline |
|---|---|---|---:|---:|---:|---|
| yield | finished_qty | ridge | 1.3956 | 1.5455 | 0.9934 | True |
| yield | yield_pct | linear | 0.0120 | 0.0129 | 0.7227 | True |
| butchery_lamb | leg_weight | linear | 0.0992 | 0.1037 | 0.9485 | True |
| butchery_lamb | loin_weight | linear | 0.0933 | 0.1084 | 0.8763 | True |
| butchery_lamb | shoulder_weight | linear | 0.0929 | 0.0980 | 0.9225 | True |
| butchery_lamb | rib_weight | linear | 0.0946 | 0.1039 | 0.8122 | True |
| butchery_lamb | mince_weight | ridge | 0.0957 | 0.0988 | 0.9063 | True |
| butchery_lamb | diced_meat_weight | ridge | 0.0985 | 0.1014 | 0.7798 | True |
| butchery_lamb | bone_weight | ridge | 0.1015 | 0.1049 | 0.9199 | True |
| butchery_lamb | waste_weight | linear | 0.1009 | 0.1008 | 0.9492 | True |
| butchery_lamb | total_saleable_weight | linear | 0.1381 | 0.1435 | 0.9936 | True |
| butchery_lamb | cutting_yield_pct | linear | 0.5920 | 0.6017 | 0.9080 | True |
| butchery_cattle | leg_weight | ridge | 1.4405 | 1.5258 | 0.8910 | True |
| butchery_cattle | loin_weight | linear | 1.5027 | 1.2902 | 0.8732 | True |
| butchery_cattle | shoulder_weight | ridge | 1.4708 | 1.5655 | 0.8508 | True |
| butchery_cattle | rib_weight | ridge | 1.5299 | 1.3574 | 0.8030 | True |
| butchery_cattle | mince_weight | ridge | 1.5316 | 1.4949 | 0.8462 | True |
| butchery_cattle | diced_meat_weight | ridge | 1.4965 | 1.4172 | 0.7625 | True |
| butchery_cattle | bone_weight | ridge | 1.5485 | 1.6195 | 0.8747 | True |
| butchery_cattle | waste_weight | linear | 1.5595 | 1.4920 | 0.9221 | True |
| butchery_cattle | total_saleable_weight | linear | 2.0689 | 2.0969 | 0.9894 | True |
| butchery_cattle | cutting_yield_pct | ridge | 0.6029 | 0.6162 | 0.9149 | True |

## Deployed predictions (physical constraints applied)

CV selects raw regressions. API quantities use reconciled parts and meat × predicted yield.

| Group | Target | Train MAE | Test MAE | Test RMSE | R² | MAPE % | Proportional baseline MAE |
|---|---|---:|---:|---:|---:|---:|---:|
| yield | finished_qty | 1.2654 | 1.4015 | 1.7698 | 0.9945 | 1.5069 | 1.4765 |
| yield | yield_pct | 0.0116 | 0.0129 | 0.0156 | 0.7227 | 1.5072 | 0.0135 |
| butchery_lamb | leg_weight | 0.0986 | 0.1036 | 0.1299 | 0.9484 | 2.6544 | 0.1098 |
| butchery_lamb | loin_weight | 0.0929 | 0.1085 | 0.1417 | 0.8756 | 4.161 | 0.1278 |
| butchery_lamb | shoulder_weight | 0.0909 | 0.0983 | 0.125 | 0.9222 | 3.3027 | 0.1121 |
| butchery_lamb | rib_weight | 0.0936 | 0.1037 | 0.1293 | 0.8124 | 6.0155 | 0.1369 |
| butchery_lamb | mince_weight | 0.095 | 0.0985 | 0.1236 | 0.9067 | 3.84 | 0.1111 |
| butchery_lamb | diced_meat_weight | 0.0973 | 0.1013 | 0.126 | 0.7795 | 6.8228 | 0.14 |
| butchery_lamb | bone_weight | 0.0993 | 0.1047 | 0.1344 | 0.9203 | 3.6882 | 0.2236 |
| butchery_lamb | waste_weight | 0.0991 | 0.1005 | 0.1269 | 0.9497 | 2.9403 | 0.2298 |
| butchery_lamb | total_saleable_weight | 0.1353 | 0.1435 | 0.1798 | 0.9936 | 0.9437 | 0.4276 |
| butchery_lamb | cutting_yield_pct | 0.6458 | 0.667 | 0.8334 | 0.8839 | 0.9437 | 1.9908 |
| butchery_cattle | leg_weight | 1.4151 | 1.5298 | 1.8946 | 0.8906 | 2.9162 | 1.74 |
| butchery_cattle | loin_weight | 1.4889 | 1.2961 | 1.6518 | 0.8725 | 3.0082 | 1.5783 |
| butchery_cattle | shoulder_weight | 1.4132 | 1.5579 | 1.9842 | 0.8521 | 3.6507 | 1.8996 |
| butchery_cattle | rib_weight | 1.5079 | 1.3553 | 1.7766 | 0.8034 | 4.6942 | 1.9156 |
| butchery_cattle | mince_weight | 1.5004 | 1.4935 | 1.8336 | 0.8468 | 3.8711 | 1.9076 |
| butchery_cattle | diced_meat_weight | 1.4691 | 1.4168 | 1.7849 | 0.763 | 5.5058 | 1.8343 |
| butchery_cattle | bone_weight | 1.5166 | 1.6158 | 2.0341 | 0.875 | 4.1282 | 3.607 |
| butchery_cattle | waste_weight | 1.5271 | 1.4975 | 1.8891 | 0.9213 | 2.7407 | 3.6221 |
| butchery_cattle | total_saleable_weight | 2.0153 | 2.097 | 2.6468 | 0.9893 | 0.9031 | 6.879 |
| butchery_cattle | cutting_yield_pct | 0.6212 | 0.6423 | 0.8025 | 0.9029 | 0.9031 | 2.0821 |

## Limitations

- Training data is `synthetic_calibrated`, not real production or carcass records.
- These metrics show that the pipeline learns the generator, not that it generalises to a real plant.
- Production EKP is a separate accounting identity and is not trained as a target.
- Real validation is blocked until real order/carcass data is provided.
- Real reference measurements exist (docs/DATA_SOURCES.md) but lack competition targets.
- No standalone validation partition: GroupKFold validates within train; fold counts in metadata.

