"""Gerçek seed verisi — AVATAR ALPHA-DUO.

`backend/migrations/0001_seed.up.sql` ve `0002_seed_more.up.sql` içindeki
gerçek üretim emirleri ve parçalama kayıtlarından derlenmiştir. ML modelleri
sentetik RNG verisi yerine bu kayıtlar üzerinde eğitilir.
"""

# Üretim emirleri: (raw_material_kg, finished_qty).
# raw_material_kg = emrin girdilerinin toplam kg; finished_qty = çıktı (mamul).
PRODUCTION_SAMPLES = [
    (107.0, 85.0),   # URT-2026-045
    (108.0, 90.0),   # URT-2026-046
    (97.0, 80.0),    # URT-2026-047
    (81.5, 70.0),    # URT-2026-048
    (106.0, 95.0),   # URT-2026-049
    (72.0, 60.0),    # URT-2026-050
    (111.0, 100.0),  # URT-2026-051
    (60.0, 55.0),    # URT-2026-052
    (100.0, 88.0),   # URT-2026-053
]

# Parçalama kayıtları: (animal_type, input_weight, [part weights]).
# Part sırası PRODUCT_CODES ile aynı: but, sırt, omuz, kaburga, kıyma, kuşbaşı, kemik.
BUTCHERY_SAMPLES = [
    ("kuzu", 25.0, [4.50, 3.00, 3.50, 2.00, 3.50, 1.75, 3.25]),   # PRC-2026-001
    ("kuzu", 22.0, [3.96, 2.64, 3.08, 1.76, 3.08, 1.54, 2.86]),   # PRC-2026-002
    ("dana", 45.0, [7.20, 5.85, 5.85, 4.05, 6.75, 3.60, 5.40]),   # PRC-2026-003
    ("kuzu", 20.0, [3.60, 2.40, 2.80, 1.60, 2.80, 1.40, 2.60]),   # PRC-2026-004
    ("kuzu", 26.0, [4.68, 3.12, 3.64, 2.08, 3.64, 1.82, 3.38]),   # PRC-2026-005
    ("dana", 50.0, [8.00, 6.50, 6.50, 4.50, 7.50, 4.00, 6.00]),   # PRC-2026-006
    ("kuzu", 24.0, [4.32, 2.88, 3.36, 1.92, 3.36, 1.68, 3.12]),   # PRC-2026-007
    ("dana", 40.0, [6.40, 5.20, 5.20, 3.60, 6.00, 3.20, 4.80]),   # PRC-2026-008
    ("kuzu", 23.0, [4.14, 2.76, 3.22, 1.84, 3.22, 1.61, 2.99]),   # PRC-2026-009
]
