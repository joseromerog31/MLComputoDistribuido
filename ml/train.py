from pathlib import Path

import joblib
import pandas as pd

from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_absolute_error, r2_score


# Rutas
BASE_DIR = Path(__file__).resolve().parent.parent

CSV_PATH = BASE_DIR / "data" / "liga_mx.csv"

MODEL_PATH = Path(__file__).resolve().parent / "model.pkl"

# Cargar
print("Cargando dataset...")

df = pd.read_csv(CSV_PATH)

print(f"Partidos encontrados: {len(df)}")

# Preparar
df["date"] = pd.to_datetime(
    df["date"],
    utc=True
)

# Trabajar cronológicamente.
df = df.sort_values("date").reset_index(drop=True)

# Proceso chistoso

# shift(1) excluye el partido actual.
# rolling(5) utiliza hasta sus 5 partidos anteriores como local.
# Se evita utilizar información futura.

df["home_avg_goals_last5"] = (
    df.groupby("home_team")["home_goals_fulltime"]
    .transform(
        lambda goals:
        goals.shift(1)
        .rolling(
            window=5,
            min_periods=1
        )
        .mean()
    )
)

# Eliminar filas inútiles
ml_data = df[
    [
        "date",
        "home_team",
        "home_avg_goals_last5",
        "home_goals_fulltime"
    ]
].dropna()


print(
    f"Partidos disponibles para ML: "
    f"{len(ml_data)}"
)

X = ml_data[
    ["home_avg_goals_last5"]
]

y = ml_data[
    "home_goals_fulltime"
]

# TRAIN / TEST

# Primer 80% = entrenamiento
# Último 20% = prueba
split_index = int(
    len(ml_data) * 0.80
)

X_train = X.iloc[:split_index]
X_test = X.iloc[split_index:]

y_train = y.iloc[:split_index]
y_test = y.iloc[split_index:]


print(
    f"Train: {len(X_train)} partidos"
)

print(
    f"Test: {len(X_test)} partidos"
)

# Entrenar
model = LinearRegression()

model.fit(
    X_train,
    y_train
)

# Evaluación
predictions = model.predict(
    X_test
)

mae = mean_absolute_error(
    y_test,
    predictions
)

r2 = r2_score(
    y_test,
    predictions
)


print("\n--- RESULTADOS ---")

print(
    f"Intercepto: "
    f"{model.intercept_:.4f}"
)

print(
    f"Coeficiente: "
    f"{model.coef_[0]:.4f}"
)

print(
    f"MAE: "
    f"{mae:.4f}"
)

print(
    f"R²: "
    f"{r2:.4f}"
)

# Guardar modelo
model_data = {
    "model": model,
    "feature": "home_avg_goals_last5",
    "mae": mae,
    "r2": r2,
}

joblib.dump(
    model_data,
    MODEL_PATH
)

print(
    f"\nModelo guardado en:"
    f"\n{MODEL_PATH}"
)