import argparse
from pathlib import Path
import time

import numpy as np
import pandas as pd
from sklearn.datasets import load_diabetes
from sklearn.metrics import mean_absolute_error, r2_score
from sklearn.model_selection import train_test_split
from tabpfn_models import REGRESSOR_MODEL_FILENAME, resolve_model_path
from tabpfn_windows_auth import windows_browser_auth_workaround
from xgboost import XGBRegressor


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Compare Prior Labs' TabPFN with XGBoost on the diabetes regression "
            "dataset."
        )
    )
    parser.add_argument("--seed", type=int, default=42)
    parser.add_argument(
        "--validation-size",
        "--test-size",
        dest="validation_size",
        type=float,
        default=0.20,
        help="Fraction of rows reserved for validation (default: 0.20).",
    )
    parser.add_argument(
        "--skip-tabpfn",
        action="store_true",
        help="Run only the XGBoost baseline.",
    )
    parser.add_argument(
        "--tabpfn-estimators",
        type=int,
        default=1,
        help="Number of TabPFN ensemble members (package default: 8).",
    )
    parser.add_argument(
        "--tabpfn-device",
        default="auto",
        help="Device such as 'auto', 'cpu', or 'cuda' (default: auto).",
    )
    parser.add_argument(
        "--tabpfn-model-path",
        type=Path,
        default=None,
        help=(
            "Optional local TabPFN checkpoint. Otherwise discover a local "
            "TabPFN-3 checkpoint before downloading."
        ),
    )
    parser.add_argument(
        "--show-predictions",
        type=int,
        default=0,
        help="Print the first N validation rows with actual and predicted values.",
    )
    return parser.parse_args()


def load_diabetes_frame() -> tuple[pd.DataFrame, pd.Series]:
    dataset = load_diabetes(as_frame=True)
    return dataset.data, dataset.target


def print_result(
    name: str,
    y_true: pd.Series,
    y_pred,
    elapsed_seconds: float,
) -> None:
    y_pred = pd.Series(y_pred, index=y_true.index, dtype="float64")
    mae = mean_absolute_error(y_true, y_pred)
    r2 = r2_score(y_true, y_pred)
    print(f"\n{name}")
    print("=" * len(name))
    print(f"MAE:     {mae:.2f}")
    print(f"R^2:     {r2:.4f}")
    print(f"Elapsed: {elapsed_seconds:.2f}s")


def print_prediction_preview(
    name: str,
    x_validation: pd.DataFrame,
    y_true: pd.Series,
    y_pred,
    limit: int,
) -> None:
    if limit <= 0:
        return

    y_pred = pd.Series(y_pred, index=y_true.index, dtype="float64")
    preview = x_validation.loc[y_true.index].copy()
    preview["actual"] = y_true.astype("float64")
    preview["predicted"] = y_pred
    preview["absolute_error"] = (preview["actual"] - preview["predicted"]).abs()

    print(f"\n{name} sample predictions:")
    print(preview.head(limit).to_string(index=False))


def run_tabpfn(
    x_train: pd.DataFrame,
    x_validation: pd.DataFrame,
    y_train: pd.Series,
    seed: int,
    n_estimators: int,
    model_path: Path | None,
    device: str,
) -> tuple[object, float]:
    from tabpfn import TabPFNRegressor

    start = time.perf_counter()
    if model_path is None:
        print("Loading the default TabPFN-3 regressor checkpoint...", flush=True)
    else:
        print(f"Loading the TabPFN checkpoint from {model_path}...", flush=True)

    regressor = TabPFNRegressor(
        model_path=model_path if model_path is not None else "auto",
        n_estimators=n_estimators,
        device=device,
        random_state=seed,
    )
    with windows_browser_auth_workaround():
        regressor.fit(x_train, y_train)
    predictions = regressor.predict(x_validation)
    if not np.isfinite(predictions).all():
        raise RuntimeError("TabPFN returned non-finite predictions.")
    elapsed = time.perf_counter() - start
    return predictions, elapsed


def run_xgboost(
    x_train: pd.DataFrame,
    x_validation: pd.DataFrame,
    y_train: pd.Series,
    seed: int,
) -> tuple[object, float]:
    regressor = XGBRegressor(
        n_estimators=200,
        max_depth=4,
        learning_rate=0.05,
        subsample=1.0,
        colsample_bytree=1.0,
        objective="reg:squarederror",
        random_state=seed,
        n_jobs=1,
    )

    start = time.perf_counter()
    regressor.fit(x_train, y_train)
    predictions = regressor.predict(x_validation)
    elapsed = time.perf_counter() - start
    return predictions, elapsed


def main() -> None:
    args = parse_args()
    x, y = load_diabetes_frame()
    x_train, x_validation, y_train, y_validation = train_test_split(
        x,
        y,
        test_size=args.validation_size,
        random_state=args.seed,
    )

    print("Dataset")
    print("=======")
    print(f"Rows:       {len(x)}")
    print(f"Features:   {x.shape[1]}")
    print(f"Training:   {len(x_train)}")
    print(f"Validation: {len(x_validation)}")
    print("Target:     disease progression score")

    if not args.skip_tabpfn:
        tabpfn_model_path = resolve_model_path(
            args.tabpfn_model_path,
            REGRESSOR_MODEL_FILENAME,
        )
        tabpfn_predictions, tabpfn_elapsed = run_tabpfn(
            x_train,
            x_validation,
            y_train,
            args.seed,
            args.tabpfn_estimators,
            tabpfn_model_path,
            args.tabpfn_device,
        )
        print_result(
            "TabPFN zero-shot",
            y_validation,
            tabpfn_predictions,
            tabpfn_elapsed,
        )
        print_prediction_preview(
            "TabPFN zero-shot",
            x_validation,
            y_validation,
            tabpfn_predictions,
            args.show_predictions,
        )

    xgboost_predictions, xgboost_elapsed = run_xgboost(
        x_train,
        x_validation,
        y_train,
        args.seed,
    )
    print_result(
        "XGBoost",
        y_validation,
        xgboost_predictions,
        xgboost_elapsed,
    )
    print_prediction_preview(
        "XGBoost",
        x_validation,
        y_validation,
        xgboost_predictions,
        args.show_predictions,
    )


if __name__ == "__main__":
    main()
