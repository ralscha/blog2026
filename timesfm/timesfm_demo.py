from __future__ import annotations

from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
from timesfm3 import ModelConfig, TimesFM3Evaluator


ROOT = Path(__file__).resolve().parent
LOCAL_CHECKPOINT = ROOT / "checkpoints" / "timesfm-3.0-pytorch"
REMOTE_CHECKPOINT = "google/timesfm-3.0-pytorch"


def load_model() -> TimesFM3Evaluator:
    checkpoint = (
        str(LOCAL_CHECKPOINT)
        if (LOCAL_CHECKPOINT / "model.safetensors").is_file()
        else REMOTE_CHECKPOINT
    )
    return TimesFM3Evaluator(
        ModelConfig(
            checkpoint_path=checkpoint,
            per_core_batch_size=4,
        )
    )


def forecast_series(
    model: TimesFM3Evaluator,
    history: np.ndarray,
    horizon: int,
    *,
    past_future_covariates: np.ndarray | None = None,
) -> tuple[np.ndarray, np.ndarray]:
    output = model.predict(
        context=history,
        horizon=horizon,
        past_future_covariates=past_future_covariates,
        return_quantiles=True,
        use_symmetric_averaging=True,
    )
    if output.forecast is None or output.quantiles is None:
        raise RuntimeError("TimesFM 3.0 did not return the requested forecasts")
    return output.forecast, output.quantiles


def ensure_output_dir(name: str) -> Path:
    output_dir = ROOT / name
    output_dir.mkdir(parents=True, exist_ok=True)
    return output_dir


def save_forecast_plot(
    *,
    output_dir: Path,
    history: np.ndarray,
    forecast: np.ndarray,
    title: str,
    history_label: str,
    forecast_label: str,
    quantiles: np.ndarray | None = None,
) -> Path:
    history_x = np.arange(len(history))
    forecast_x = np.arange(len(history), len(history) + len(forecast))

    plt.figure(figsize=(12, 6))
    plt.plot(history_x, history, label=history_label, linewidth=2)
    plt.plot(forecast_x, forecast, label=forecast_label, linestyle="--", linewidth=2)

    if quantiles is not None and quantiles.shape[-1] >= 9:
        lower_band = quantiles[:, 0]
        upper_band = quantiles[:, 8]
        plt.fill_between(
            forecast_x,
            lower_band,
            upper_band,
            alpha=0.2,
            label="10%-90% band",
        )

    plt.axvline(len(history) - 1, color="black", linewidth=1, linestyle=":")
    plt.title(title)
    plt.legend()
    plt.tight_layout()

    output_path = output_dir / "forecast.png"
    plt.savefig(output_path, dpi=144)
    plt.close()
    return output_path


def quantile_frame(quantiles: np.ndarray) -> pd.DataFrame:
    columns = [
        "p10",
        "p20",
        "p30",
        "p40",
        "p50",
        "p60",
        "p70",
        "p80",
        "p90",
    ]
    if quantiles.ndim != 2 or quantiles.shape[1] != len(columns):
        raise ValueError(
            "Expected TimesFM 3.0 quantiles with shape (horizon, 9), "
            f"got {quantiles.shape}"
        )
    return pd.DataFrame(quantiles, columns=columns)
