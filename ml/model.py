import numpy as np
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_absolute_error, mean_squared_error


def train_yield(features, y):
    """Train a linear regression model. X: (n, k) ndarray, y: (n,) ndarray."""
    model = LinearRegression()
    model.fit(features, y)
    return model


def predict(model, features):
    return model.predict(features)


def baseline_mean(y):
    """Reference prediction: the mean of all observed values."""
    return float(np.mean(y))


def evaluate(y_true, y_pred):
    """Return mean absolute error and root mean squared error."""
    mae = mean_absolute_error(y_true, y_pred)
    rmse = np.sqrt(mean_squared_error(y_true, y_pred))
    return {"mae": float(mae), "rmse": float(rmse)}
