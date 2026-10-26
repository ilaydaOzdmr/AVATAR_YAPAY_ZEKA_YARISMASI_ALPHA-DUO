import pytest

from model import baseline_mean, evaluate, predict, train_yield


def test_baseline_mean():
    assert baseline_mean([10, 20, 30]) == 20.0


def test_evaluate():
    m = evaluate([1, 2, 3], [1, 2, 3])
    assert m["mae"] == 0.0 and m["rmse"] == 0.0


def test_train_predict_shape():
    X = [[1], [2], [3], [4]]
    y = [2, 4, 6, 8]
    model = train_yield(X, y)
    assert predict(model, [[5]])[0] == pytest.approx(10.0, abs=1e-6)
