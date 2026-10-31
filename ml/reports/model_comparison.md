# Model comparison

All models were selected by grouped cross-validation MAE. The test set was used only for reporting, never for selection or hyperparameter tuning.

> Data origin: `synthetic_calibrated`. These are not real field results.

## yield

### finished_qty

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 1.3956 | 0.2541 | 1.5455 | 1.9530 | 0.9934 | 1.2775 | 1.64 |
| linear | 1.4015 | 0.2543 | 1.5320 | 1.9263 | 0.9935 | 1.2934 | 1.63 |
| random_forest | 1.8979 | 0.1078 | 2.2027 | 2.8094 | 0.9862 | 1.8309 | 2.41 |
| dummy_mean | 21.4447 | 1.2840 | 20.3131 | 24.0552 | -0.0081 | 20.4654 | 25.18 |

Selected: **ridge**; beats mean baseline: **True**.

### yield_pct

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 0.0120 | 0.0022 | 0.0129 | 0.0156 | 0.7227 | 0.0118 | 1.51 |
| ridge | 0.0120 | 0.0022 | 0.0129 | 0.0156 | 0.7224 | 0.0118 | 1.51 |
| random_forest | 0.0126 | 0.0017 | 0.0128 | 0.0156 | 0.7210 | 0.0119 | 1.49 |
| dummy_mean | 0.0198 | 0.0012 | 0.0231 | 0.0296 | -0.0036 | 0.0199 | 2.66 |

Selected: **linear**; beats mean baseline: **True**.

## butchery_lamb

### leg_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 0.0992 | 0.0045 | 0.1037 | 0.1298 | 0.9485 | 0.0872 | 2.66 |
| ridge | 0.0992 | 0.0045 | 0.1037 | 0.1298 | 0.9485 | 0.0872 | 2.66 |
| random_forest | 0.1099 | 0.0025 | 0.1043 | 0.1332 | 0.9458 | 0.0828 | 2.69 |
| dummy_mean | 0.4464 | 0.0290 | 0.4573 | 0.5778 | -0.0205 | 0.3613 | 12.05 |

Selected: **linear**; beats mean baseline: **True**.

### loin_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 0.0933 | 0.0060 | 0.1084 | 0.1413 | 0.8763 | 0.0856 | 4.16 |
| ridge | 0.0933 | 0.0060 | 0.1083 | 0.1412 | 0.8765 | 0.0858 | 4.16 |
| random_forest | 0.1023 | 0.0099 | 0.1151 | 0.1501 | 0.8604 | 0.0946 | 4.43 |
| dummy_mean | 0.3271 | 0.0212 | 0.3226 | 0.4054 | -0.0178 | 0.2650 | 12.80 |

Selected: **linear**; beats mean baseline: **True**.

### shoulder_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 0.0929 | 0.0086 | 0.0980 | 0.1247 | 0.9225 | 0.0759 | 3.29 |
| ridge | 0.0929 | 0.0086 | 0.0980 | 0.1247 | 0.9225 | 0.0760 | 3.29 |
| random_forest | 0.1031 | 0.0062 | 0.1048 | 0.1328 | 0.9122 | 0.0879 | 3.51 |
| dummy_mean | 0.3614 | 0.0218 | 0.3599 | 0.4534 | -0.0239 | 0.2920 | 12.09 |

Selected: **linear**; beats mean baseline: **True**.

### rib_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 0.0946 | 0.0056 | 0.1039 | 0.1293 | 0.8122 | 0.0830 | 6.02 |
| ridge | 0.0946 | 0.0056 | 0.1039 | 0.1294 | 0.8121 | 0.0832 | 6.03 |
| random_forest | 0.1057 | 0.0116 | 0.1066 | 0.1360 | 0.7924 | 0.0869 | 6.26 |
| dummy_mean | 0.2308 | 0.0141 | 0.2399 | 0.3022 | -0.0253 | 0.2155 | 14.24 |

Selected: **linear**; beats mean baseline: **True**.

### mince_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 0.0957 | 0.0065 | 0.0988 | 0.1239 | 0.9063 | 0.0829 | 3.85 |
| linear | 0.0958 | 0.0065 | 0.0983 | 0.1234 | 0.9069 | 0.0838 | 3.83 |
| random_forest | 0.1049 | 0.0083 | 0.1015 | 0.1308 | 0.8955 | 0.0812 | 3.97 |
| dummy_mean | 0.3155 | 0.0264 | 0.3221 | 0.4110 | -0.0316 | 0.2546 | 12.59 |

Selected: **ridge**; beats mean baseline: **True**.

### diced_meat_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 0.0985 | 0.0090 | 0.1014 | 0.1259 | 0.7798 | 0.0866 | 6.83 |
| linear | 0.0987 | 0.0089 | 0.1011 | 0.1261 | 0.7793 | 0.0846 | 6.82 |
| random_forest | 0.1049 | 0.0131 | 0.1095 | 0.1372 | 0.7388 | 0.0971 | 7.41 |
| dummy_mean | 0.2193 | 0.0161 | 0.2131 | 0.2697 | -0.0097 | 0.1604 | 14.69 |

Selected: **ridge**; beats mean baseline: **True**.

### bone_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 0.1015 | 0.0044 | 0.1049 | 0.1347 | 0.9199 | 0.0901 | 3.69 |
| linear | 0.1016 | 0.0039 | 0.1046 | 0.1338 | 0.9210 | 0.0909 | 3.70 |
| random_forest | 0.1189 | 0.0064 | 0.1212 | 0.1578 | 0.8901 | 0.1024 | 4.26 |
| dummy_mean | 0.3576 | 0.0234 | 0.3788 | 0.4843 | -0.0351 | 0.3235 | 13.55 |

Selected: **ridge**; beats mean baseline: **True**.

### waste_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 0.1009 | 0.0031 | 0.1008 | 0.1275 | 0.9492 | 0.0857 | 2.95 |
| ridge | 0.1009 | 0.0031 | 0.1009 | 0.1277 | 0.9491 | 0.0858 | 2.95 |
| random_forest | 0.1229 | 0.0096 | 0.1325 | 0.1680 | 0.9118 | 0.1103 | 3.85 |
| dummy_mean | 0.4258 | 0.0118 | 0.4459 | 0.5753 | -0.0341 | 0.3565 | 13.04 |

Selected: **linear**; beats mean baseline: **True**.

### total_saleable_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 0.1381 | 0.0033 | 0.1435 | 0.1794 | 0.9936 | 0.1250 | 0.94 |
| ridge | 0.1381 | 0.0033 | 0.1435 | 0.1793 | 0.9936 | 0.1249 | 0.94 |
| random_forest | 0.2263 | 0.0216 | 0.2086 | 0.2716 | 0.9853 | 0.1577 | 1.37 |
| dummy_mean | 1.7869 | 0.1185 | 1.7723 | 2.2702 | -0.0244 | 1.4245 | 11.82 |

Selected: **linear**; beats mean baseline: **True**.

### cutting_yield_pct

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 0.5920 | 0.0265 | 0.6017 | 0.7417 | 0.9080 | 0.5145 | 0.85 |
| ridge | 0.5920 | 0.0264 | 0.6017 | 0.7418 | 0.9080 | 0.5149 | 0.85 |
| random_forest | 0.6329 | 0.0246 | 0.6411 | 0.7777 | 0.8989 | 0.5512 | 0.90 |
| dummy_mean | 2.0575 | 0.0895 | 1.9908 | 2.4495 | -0.0030 | 1.7580 | 2.82 |

Selected: **linear**; beats mean baseline: **True**.

## butchery_cattle

### leg_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 1.4405 | 0.0571 | 1.5258 | 1.8907 | 0.8910 | 1.3427 | 2.91 |
| linear | 1.4408 | 0.0572 | 1.5267 | 1.8915 | 0.8910 | 1.3496 | 2.91 |
| random_forest | 1.5836 | 0.0522 | 1.6251 | 2.0200 | 0.8756 | 1.3928 | 3.09 |
| dummy_mean | 4.8838 | 0.1171 | 4.4853 | 5.7304 | -0.0008 | 4.0143 | 8.60 |

Selected: **ridge**; beats mean baseline: **True**.

### loin_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 1.5027 | 0.0487 | 1.2902 | 1.6472 | 0.8732 | 1.1107 | 3.00 |
| ridge | 1.5027 | 0.0487 | 1.2901 | 1.6472 | 0.8732 | 1.1107 | 3.00 |
| random_forest | 1.6118 | 0.0335 | 1.4603 | 1.7843 | 0.8512 | 1.2202 | 3.41 |
| dummy_mean | 4.1647 | 0.1462 | 3.6836 | 4.6303 | -0.0019 | 3.1015 | 8.74 |

Selected: **linear**; beats mean baseline: **True**.

### shoulder_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 1.4708 | 0.1372 | 1.5655 | 1.9924 | 0.8508 | 1.3225 | 3.67 |
| linear | 1.4731 | 0.1398 | 1.5548 | 1.9796 | 0.8527 | 1.3629 | 3.64 |
| random_forest | 1.6650 | 0.1905 | 1.6121 | 2.1355 | 0.8286 | 1.2981 | 3.75 |
| dummy_mean | 4.0683 | 0.1437 | 3.9824 | 5.1590 | -0.0001 | 3.2953 | 9.42 |

Selected: **ridge**; beats mean baseline: **True**.

### rib_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 1.5299 | 0.0403 | 1.3574 | 1.7786 | 0.8030 | 1.0199 | 4.70 |
| linear | 1.5301 | 0.0430 | 1.3540 | 1.7733 | 0.8042 | 1.0275 | 4.69 |
| random_forest | 1.5903 | 0.0638 | 1.4726 | 1.9550 | 0.7620 | 1.1264 | 5.05 |
| dummy_mean | 3.3175 | 0.1375 | 3.1911 | 4.0140 | -0.0034 | 2.7418 | 11.08 |

Selected: **ridge**; beats mean baseline: **True**.

### mince_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 1.5316 | 0.1147 | 1.4949 | 1.8368 | 0.8462 | 1.2711 | 3.87 |
| linear | 1.5322 | 0.1160 | 1.4918 | 1.8313 | 0.8472 | 1.2522 | 3.87 |
| random_forest | 1.6707 | 0.1337 | 1.6601 | 2.0338 | 0.8115 | 1.3875 | 4.29 |
| dummy_mean | 3.8694 | 0.0860 | 3.7826 | 4.6841 | -0.0000 | 3.2184 | 9.73 |

Selected: **ridge**; beats mean baseline: **True**.

### diced_meat_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 1.4965 | 0.0842 | 1.4172 | 1.7868 | 0.7625 | 1.2207 | 5.50 |
| linear | 1.4971 | 0.0837 | 1.4173 | 1.7845 | 0.7631 | 1.2641 | 5.51 |
| random_forest | 1.6094 | 0.1314 | 1.4283 | 1.8633 | 0.7417 | 1.1854 | 5.49 |
| dummy_mean | 3.1155 | 0.0729 | 2.7847 | 3.6691 | -0.0016 | 2.2035 | 10.70 |

Selected: **ridge**; beats mean baseline: **True**.

### bone_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 1.5485 | 0.0990 | 1.6195 | 2.0362 | 0.8747 | 1.3064 | 4.14 |
| linear | 1.5499 | 0.1043 | 1.6116 | 2.0315 | 0.8753 | 1.2964 | 4.12 |
| random_forest | 1.7844 | 0.1267 | 1.9286 | 2.3491 | 0.8333 | 1.6821 | 4.92 |
| dummy_mean | 4.2448 | 0.1842 | 4.6805 | 5.7537 | -0.0003 | 4.2990 | 12.23 |

Selected: **ridge**; beats mean baseline: **True**.

### waste_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 1.5595 | 0.1283 | 1.4920 | 1.8794 | 0.9221 | 1.2533 | 2.73 |
| ridge | 1.5595 | 0.1284 | 1.4919 | 1.8792 | 0.9221 | 1.2532 | 2.73 |
| random_forest | 1.8653 | 0.1652 | 1.6560 | 2.0501 | 0.9073 | 1.4986 | 3.01 |
| dummy_mean | 5.4107 | 0.2755 | 5.5461 | 6.7321 | -0.0000 | 4.8805 | 10.27 |

Selected: **linear**; beats mean baseline: **True**.

### total_saleable_weight

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| linear **selected** | 2.0689 | 0.1200 | 2.0969 | 2.6446 | 0.9894 | 1.6792 | 0.90 |
| ridge | 2.0689 | 0.1203 | 2.0969 | 2.6451 | 0.9894 | 1.6778 | 0.90 |
| random_forest | 3.2844 | 0.2422 | 3.0395 | 3.9804 | 0.9759 | 2.5775 | 1.29 |
| dummy_mean | 21.6525 | 0.4505 | 20.2062 | 25.6464 | -0.0009 | 17.9985 | 8.73 |

Selected: **linear**; beats mean baseline: **True**.

### cutting_yield_pct

| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |
|---|---:|---:|---:|---:|---:|---:|---:|
| ridge **selected** | 0.6029 | 0.0301 | 0.6162 | 0.7511 | 0.9149 | 0.5236 | 0.87 |
| linear | 0.6030 | 0.0298 | 0.6161 | 0.7512 | 0.9149 | 0.5251 | 0.87 |
| random_forest | 0.6173 | 0.0363 | 0.6238 | 0.7777 | 0.9088 | 0.5631 | 0.88 |
| dummy_mean | 2.0941 | 0.1068 | 2.0821 | 2.5768 | -0.0012 | 1.8416 | 2.95 |

Selected: **ridge**; beats mean baseline: **True**.

