from __future__ import annotations

import argparse
import sys
from pathlib import Path
from urllib.error import URLError
from urllib.request import urlopen

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
from statsmodels.tsa.seasonal import seasonal_decompose

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from timesfm_demo import ensure_output_dir, forecast_series, load_model, quantile_frame


DEFAULT_CSV_PATH = Path(__file__).resolve().parent / "train.csv"
DEFAULT_CSV_URL = (
    "https://raw.githubusercontent.com/"
    "gagandeepsinghkhanuja/Walmart-Sales-Forecasting/master/train.csv"
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Run seasonal decomposition and a TimesFM forecast on Walmart weekly sales."
    )
    parser.add_argument(
        "--csv",
        type=Path,
        default=DEFAULT_CSV_PATH,
        help="Path to the Walmart train.csv file",
    )
    parser.add_argument(
        "--horizon",
        type=int,
        default=12,
        help="Forecast horizon in weeks",
    )
    return parser.parse_args()


def ensure_default_csv(csv_path: Path) -> Path:
    if csv_path.exists() or csv_path != DEFAULT_CSV_PATH:
        return csv_path

    try:
        with urlopen(DEFAULT_CSV_URL) as response:
            csv_path.write_bytes(response.read())
    except URLError as exc:
        raise FileNotFoundError(
            f"Could not find Walmart data at {csv_path} and failed to download it from {DEFAULT_CSV_URL}. "
            "Pass --csv with a local train.csv file instead."
        ) from exc

    return csv_path


def load_weekly_sales(csv_path: Path) -> pd.Series:
    csv_path = ensure_default_csv(csv_path)

    if not csv_path.exists():
        raise FileNotFoundError(
            f"Could not find Walmart data at {csv_path}. Download train.csv and pass --csv."
        )

    frame = pd.read_csv(csv_path, usecols=["Date", "Weekly_Sales"])
    frame["Date"] = pd.to_datetime(frame["Date"])
    weekly_sales = frame.groupby("Date")["Weekly_Sales"].sum().sort_index()

    if len(weekly_sales) < 104:
        raise RuntimeError(
            "Expected at least two years of weekly sales for a 52-week decomposition"
        )

    return weekly_sales


def save_decomposition_plot(series: pd.Series, output_dir: Path) -> Path:
    analysis = seasonal_decompose(
        series, model="additive", period=52, extrapolate_trend="period"
    )
    figure = analysis.plot()
    figure.set_size_inches(12, 9)
    figure.tight_layout()

    output_path = output_dir / "decomposition.png"
    figure.savefig(output_path, dpi=144)
    plt.close(figure)
    return output_path


def save_forecast_plot(
    *,
    history_dates: pd.DatetimeIndex,
    history_values: np.ndarray,
    forecast_dates: pd.DatetimeIndex,
    forecast_values: np.ndarray,
    quantiles: np.ndarray,
    horizon: int,
    output_dir: Path,
) -> Path:
    plt.figure(figsize=(12, 6))
    plt.plot(
        history_dates, history_values, label="Historical weekly sales", linewidth=2
    )
    plt.plot(
        forecast_dates,
        forecast_values,
        label=f"TimesFM {horizon}-week forecast",
        linestyle="--",
        linewidth=2,
    )

    if quantiles.shape[-1] >= 9:
        plt.fill_between(
            forecast_dates,
            quantiles[:, 0],
            quantiles[:, 8],
            alpha=0.2,
            label="10%-90% band",
        )

    plt.title("Walmart weekly sales forecast with TimesFM 3.0")
    plt.xlabel("Date")
    plt.ylabel("Total weekly sales")
    plt.legend()
    plt.tight_layout()

    output_path = output_dir / "forecast.png"
    plt.savefig(output_path, dpi=144)
    plt.close()
    return output_path


def main() -> None:
    args = parse_args()
    weekly_sales = load_weekly_sales(args.csv)

    output_dir = ensure_output_dir("walmart-sales")
    decomposition_path = save_decomposition_plot(weekly_sales, output_dir)

    history = weekly_sales.to_numpy(dtype="float32")
    model = load_model()
    point_forecast, quantile_forecast = forecast_series(
        model, history, horizon=args.horizon
    )

    last_date = weekly_sales.index[-1]
    forecast_dates = pd.date_range(
        start=last_date + pd.Timedelta(weeks=1),
        periods=args.horizon,
        freq="W-FRI",
    )

    forecast_path = save_forecast_plot(
        history_dates=weekly_sales.index,
        history_values=history,
        forecast_dates=forecast_dates,
        forecast_values=point_forecast,
        quantiles=quantile_forecast,
        horizon=args.horizon,
        output_dir=output_dir,
    )

    quantiles = quantile_frame(quantile_forecast)
    quantiles.insert(0, "date", forecast_dates.strftime("%Y-%m-%d"))
    quantiles.to_csv(output_dir / "forecast_quantiles.csv", index=False)

    print(f"Saved decomposition plot to {decomposition_path}")
    print(f"Saved forecast plot to {forecast_path}")
    print(quantiles.to_string(index=False))


if __name__ == "__main__":
    main()
