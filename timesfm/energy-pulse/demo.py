from __future__ import annotations

import io
import ssl
import sys
import urllib.request
from pathlib import Path

import numpy as np
import pandas as pd

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from timesfm_demo import (
    ensure_output_dir,
    forecast_series,
    load_model,
    quantile_frame,
    save_forecast_plot,
)


URL = "https://data.open-power-system-data.org/time_series/2020-10-06/time_series_60min_singleindex.csv"
COLUMN = "DE_load_actual_entsoe_transparency"


def load_energy_series() -> np.ndarray:
    try:
        frame = pd.read_csv(URL, index_col=0, parse_dates=True)
        return frame[COLUMN].ffill().dropna().to_numpy(dtype="float32")
    except Exception as exc:
        print(f"Warning: TLS-verified download failed ({exc}); retrying with TLS verification disabled.")
        context = ssl._create_unverified_context()
        with urllib.request.urlopen(URL, context=context) as response:
            frame = pd.read_csv(io.BytesIO(response.read()), index_col=0, parse_dates=True)
        return frame[COLUMN].ffill().dropna().to_numpy(dtype="float32")


def main() -> None:
    load = load_energy_series()

    history = load[-24 * 28 :]
    model = load_model()
    point_forecast, quantile_forecast = forecast_series(
        model, history, horizon=24 * 7
    )

    output_dir = ensure_output_dir("energy-pulse")
    plot_path = save_forecast_plot(
        output_dir=output_dir,
        history=history[-24 * 7 :],
        forecast=point_forecast,
        quantiles=quantile_forecast,
        title="German electricity load forecast with TimesFM 3.0",
        history_label="Last 7 days of load",
        forecast_label="Next 7 days forecast",
    )

    quantiles = quantile_frame(quantile_forecast)
    quantiles.insert(0, "hour", range(1, len(quantiles) + 1))
    quantiles.to_csv(output_dir / "forecast_quantiles.csv", index=False)

    print(f"Saved plot to {plot_path}")
    print(quantiles.head(24).to_string(index=False))


if __name__ == "__main__":
    main()
