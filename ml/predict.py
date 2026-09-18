from pathlib import Path
import json
import sys

import joblib
import pandas as pd


MODEL_PATH = Path(__file__).resolve().parent / "model.pkl"
FEATURE_NAME = "home_avg_goals_last5"


def predict_batch(records):
    if not records:
        return []

    model_data = joblib.load(MODEL_PATH)

    model = model_data["model"]
    feature = model_data.get("feature")

    if feature != FEATURE_NAME:
        raise ValueError(
            f"El modelo espera la feature '{feature}', "
            f"pero predict.py utiliza '{FEATURE_NAME}'."
        )

    dataframe = pd.DataFrame(
        {
            FEATURE_NAME: [
                record[FEATURE_NAME]
                for record in records
            ]
        }
    )

    predictions = model.predict(dataframe)

    results = []

    for record, prediction in zip(records, predictions):
        results.append(
            {
                "id": record["id"],
                "prediction": float(prediction),
            }
        )

    return results


def main():
    try:
        data = json.load(sys.stdin)

        records = data.get("records", [])

        predictions = predict_batch(records)

        print(
            json.dumps(
                {
                    "predictions": predictions
                }
            )
        )

    except Exception as error:
        print(
            json.dumps(
                {
                    "error": str(error)
                }
            )
        )
        sys.exit(1)


if __name__ == "__main__":
    main()