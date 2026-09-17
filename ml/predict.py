from pathlib import Path
import sys

import joblib


MODEL_PATH = (
    Path(__file__).resolve().parent
    / "model.pkl"
)


def predict(avg_goals: float) -> float:

    model_data = joblib.load(
        MODEL_PATH
    )

    model = model_data["model"]

    prediction = model.predict(
        [[avg_goals]]
    )[0]

    return float(prediction)


if __name__ == "__main__":

    if len(sys.argv) != 2:

        print(
            "Uso: python3 ml/predict.py "
            "<promedio_goles>"
        )

        sys.exit(1)

    avg_goals = float(
        sys.argv[1]
    )

    result = predict(
        avg_goals
    )

    print(
        f"Promedio anterior: "
        f"{avg_goals}"
    )

    print(
        f"Predicción de goles: "
        f"{result:.2f}"
    )