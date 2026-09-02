from __future__ import annotations

import sys
from pathlib import Path

import numpy as np

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from timesfm_demo import (
    ensure_output_dir,
    forecast_series,
    load_model,
    quantile_frame,
    save_forecast_plot,
)


HORIZON = 48


def build_series_and_covariate() -> tuple[np.ndarray, np.ndarray]:
    hours = np.arange(14 * 24)
    daily_cycle = 40 + 8 * np.sin(2 * np.pi * hours / 24)
    weekly_cycle = 4 * np.sin(2 * np.pi * hours / (24 * 7))
    history = daily_cycle + weekly_cycle

    event_active = np.zeros(len(history) + HORIZON, dtype=np.float32)
    event_start = len(history) - 72
    event_active[event_start : len(history)] = 1

    history[event_start : event_start + 72] += 28
    history[event_start + 24 : event_start + 48] += 12
    return history.astype(np.float32), event_active


def main() -> None:
    history, event_active = build_series_and_covariate()
    model = load_model()
    point_forecast, quantile_forecast = forecast_series(
        model,
        history,
        horizon=HORIZON,
        past_future_covariates=event_active,
    )

    output_dir = ensure_output_dir("taylor-swift-xreg")
    plot_path = save_forecast_plot(
        output_dir=output_dir,
        history=history,
        forecast=point_forecast,
        quantiles=quantile_forecast,
        title="Synthetic event shock forecast with a TimesFM 3.0 covariate",
        history_label="Hourly demand history",
        forecast_label="48-hour covariate forecast",
    )

    quantiles = quantile_frame(quantile_forecast)
    quantiles.insert(0, "event_active", event_active[len(history) :])
    quantiles.insert(0, "step", np.arange(1, len(quantiles) + 1))
    quantiles.to_csv(output_dir / "forecast_quantiles.csv", index=False)

    print(f"Saved plot to {plot_path}")
    print(quantiles.head(10).to_string(index=False))


if __name__ == "__main__":
    main()
