from __future__ import annotations

from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import timesfm


ROOT = Path(__file__).resolve().parent


def load_model(
    max_context: int,
    max_horizon: int,
    *,
    return_backcast: bool = False,
) -> timesfm.TimesFM_2p5_200M_torch:
    model = timesfm.TimesFM_2p5_200M_torch.from_pretrained(
        "google/timesfm-2.5-200m-pytorch"
    )
    model.compile(
        timesfm.ForecastConfig(
            max_context=max_context,
            max_horizon=max_horizon,
            normalize_inputs=True,
            use_continuous_quantile_head=True,
            force_flip_invariance=True,
            fix_quantile_crossing=True,
            return_backcast=return_backcast,
        )
    )
    return model


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

    if quantiles is not None and quantiles.shape[-1] >= 10:
        lower_band = quantiles[:, 1]
        upper_band = quantiles[:, 9]
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
        "mean",
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
    return pd.DataFrame(quantiles, columns=columns[: quantiles.shape[1]])
