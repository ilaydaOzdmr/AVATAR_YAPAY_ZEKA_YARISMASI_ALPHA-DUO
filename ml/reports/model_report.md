# Model report

Model version: `1.0.0` · Status: `synthetic_only`

## What was trained

| Group | Targets | Train rows | Test rows | Validation |
|---|---|---:|---:|---|
| yield | 2 | 394 | 86 | GroupShuffleSplit(test_size=0.2, seed=42) + GroupKFold(5) |
| butchery_lamb | 10 | 592 | 208 | GroupShuffleSplit(test_size=0.2, seed=42) + GroupKFold(5) |
| butchery_cattle | 10 | 600 | 200 | GroupShuffleSplit(test_size=0.2, seed=42) + GroupKFold(5) |

## Selected models

| Group | Target | Model | CV MAE | Test MAE | Test R² | Beats baseline |
|---|---|---|---:|---:|---:|---|
| yield | finished_qty | lasso | 1.4014 | 1.4535 | 0.9945 | True |
| yield | yield_pct | linear | 0.0118 | 0.0131 | 0.6882 | True |
| butchery_lamb | leg_weight | ridge | 0.0988 | 0.1038 | 0.9490 | True |
| butchery_lamb | loin_weight | lasso | 0.0931 | 0.1082 | 0.8767 | True |
| butchery_lamb | shoulder_weight | linear | 0.0926 | 0.0986 | 0.9221 | True |
| butchery_lamb | rib_weight | lasso | 0.0943 | 0.1038 | 0.8123 | True |
| butchery_lamb | mince_weight | lasso | 0.0945 | 0.0985 | 0.9073 | True |
| butchery_lamb | diced_meat_weight | lasso | 0.0979 | 0.1038 | 0.7712 | True |
| butchery_lamb | bone_weight | lasso | 0.1018 | 0.1047 | 0.9211 | True |
| butchery_lamb | waste_weight | lasso | 0.1005 | 0.1010 | 0.9489 | True |
| butchery_lamb | total_saleable_weight | lasso | 0.1378 | 0.1433 | 0.9936 | True |
| butchery_lamb | cutting_yield_pct | lasso | 0.5905 | 0.5985 | 0.9088 | True |
| butchery_cattle | leg_weight | lasso | 1.4400 | 1.5241 | 0.8914 | True |
| butchery_cattle | loin_weight | lasso | 1.5057 | 1.3024 | 0.8731 | True |
| butchery_cattle | shoulder_weight | lasso | 1.4553 | 1.5651 | 0.8496 | True |
| butchery_cattle | rib_weight | lasso | 1.5187 | 1.3753 | 0.7966 | True |
| butchery_cattle | mince_weight | lasso | 1.5223 | 1.5134 | 0.8426 | True |
| butchery_cattle | diced_meat_weight | lasso | 1.4992 | 1.4101 | 0.7653 | True |
| butchery_cattle | bone_weight | ridge | 1.5505 | 1.6183 | 0.8735 | True |
| butchery_cattle | waste_weight | ridge | 1.5513 | 1.4996 | 0.9215 | True |
| butchery_cattle | total_saleable_weight | linear | 2.0682 | 2.1227 | 0.9891 | True |
| butchery_cattle | cutting_yield_pct | ridge | 0.6034 | 0.6244 | 0.9129 | True |

## Limitations

- Training data is `synthetic_calibrated`, not real production or carcass records.
- These metrics show that the pipeline learns the generator, not that it generalises to a real plant.
- Production EKP is a separate accounting identity and is not trained as a target.
- Real validation is blocked until real order/carcass data is provided.

