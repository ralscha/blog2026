from __future__ import annotations

import sys
from pathlib import Path

import yfinance as yf

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from timesfm_demo import (
    ensure_output_dir,
    load_model,
    quantile_frame,
    save_forecast_plot,
)


def main() -> None:
    prices = yf.download(
        "NVDA", period="2y", interval="1d", auto_adjust=True, progress=False
    )
    close = prices["Close"]
    if getattr(close, "ndim", 1) != 1:
        close = close.squeeze("columns")

    close_series = close.dropna().to_numpy(dtype="float32")

    if len(close_series) < 200:
        raise RuntimeError("Expected at least 200 closing prices from Yahoo Finance")

    history = close_series[-512:]
    model = load_model(max_context=512, max_horizon=30)
    point_forecast, quantile_forecast = model.forecast(horizon=30, inputs=[history])

    output_dir = ensure_output_dir("market-vibe")
    plot_path = save_forecast_plot(
        output_dir=output_dir,
        history=history,
        forecast=point_forecast[0],
        quantiles=quantile_forecast[0],
        title="NVDA closing price forecast with TimesFM 2.5",
        history_label="Adjusted close",
        forecast_label="30-trading-day forecast",
    )

    quantiles = quantile_frame(quantile_forecast[0])
    quantiles.insert(0, "step", range(1, len(quantiles) + 1))
    quantiles.to_csv(output_dir / "forecast_quantiles.csv", index=False)

    print(f"Saved plot to {plot_path}")
    print(quantiles.head(10).to_string(index=False))


if __name__ == "__main__":
    main()
