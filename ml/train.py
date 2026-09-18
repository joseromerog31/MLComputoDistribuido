from pathlib import Path
import os

import joblib
import pandas as pd
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_absolute_error, r2_score
from sqlalchemy import create_engine


MODEL_PATH = Path(__file__).resolve().parent / "model.pkl"

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgresql+psycopg://postgres:postgres@localhost:5432/postgres",
)


def load_data():
    engine = create_engine(DATABASE_URL)

    query = """
        SELECT
            id,
            date,
            home_avg_goals_last5::DOUBLE PRECISION AS home_avg_goals_last5,
            home_goals_fulltime
        FROM ml_features
        WHERE home_avg_goals_last5 IS NOT NULL
          AND home_goals_fulltime IS NOT NULL
        ORDER BY date, id;
    """

    dataframe = pd.read_sql(query, engine)

    engine.dispose()

    return dataframe


def train_model(dataframe):
    X = dataframe[["home_avg_goals_last5"]]
    y = dataframe["home_goals_fulltime"]

    split_index = int(len(dataframe) * 0.8)

    X_train = X.iloc[:split_index]
    X_test = X.iloc[split_index:]

    y_train = y.iloc[:split_index]
    y_test = y.iloc[split_index:]

    model = LinearRegression()
    model.fit(X_train, y_train)

    predictions = model.predict(X_test)

    mae = mean_absolute_error(y_test, predictions)
    r2 = r2_score(y_test, predictions)

    return model, mae, r2, len(X_train), len(X_test)


def save_model(model, mae, r2):
    model_data = {
        "model": model,
        "feature": "home_avg_goals_last5",
        "mae": mae,
        "r2": r2,
    }

    joblib.dump(model_data, MODEL_PATH)


def main():
    dataframe = load_data()

    if dataframe.empty:
        raise ValueError("No hay datos disponibles para entrenar el modelo.")

    model, mae, r2, train_size, test_size = train_model(dataframe)

    print(f"Registros disponibles: {len(dataframe)}")
    print(f"Entrenamiento: {train_size}")
    print(f"Prueba: {test_size}")
    print()
    print(f"Intercepto: {model.intercept_:.4f}")
    print(f"Coeficiente: {model.coef_[0]:.4f}")
    print(f"MAE: {mae:.4f}")
    print(f"R²: {r2:.4f}")
    print()
    print(f"Modelo guardado en: {MODEL_PATH}")

    save_model(model, mae, r2)


if __name__ == "__main__":
    main()