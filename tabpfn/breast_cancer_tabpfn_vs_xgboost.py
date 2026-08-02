import argparse
from pathlib import Path
import time

import numpy as np
import pandas as pd
from sklearn.datasets import load_breast_cancer
from sklearn.metrics import accuracy_score, classification_report, confusion_matrix
from sklearn.model_selection import train_test_split
from tabpfn_models import CLASSIFIER_MODEL_FILENAME, resolve_model_path
from tabpfn_windows_auth import windows_browser_auth_workaround
from xgboost import XGBClassifier


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Compare Prior Labs' TabPFN with XGBoost on the breast cancer dataset."
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
        help="Print the first N validation rows with actual and predicted labels.",
    )
    return parser.parse_args()


def load_breast_cancer_frame() -> tuple[pd.DataFrame, pd.Series, list[str]]:
    dataset = load_breast_cancer(as_frame=True)
    return dataset.data, dataset.target, list(dataset.target_names)


def print_result(
    name: str,
    y_true: pd.Series,
    y_pred,
    target_names: list[str],
    elapsed_seconds: float,
) -> None:
    y_pred = pd.Series(y_pred, index=y_true.index).astype(y_true.dtype)
    accuracy = accuracy_score(y_true, y_pred)
    print(f"\n{name}")
    print("=" * len(name))
    print(f"Accuracy: {accuracy:.4f}")
    print(f"Elapsed:  {elapsed_seconds:.2f}s")
    print("\nConfusion matrix:")
    print(confusion_matrix(y_true, y_pred))
    print("\nClassification report:")
    print(
        classification_report(
            y_true,
            y_pred,
            target_names=target_names,
            zero_division=0,
        )
    )


def print_prediction_preview(
    name: str,
    x_validation: pd.DataFrame,
    y_true: pd.Series,
    y_pred,
    target_names: list[str],
    limit: int,
) -> None:
    if limit <= 0:
        return

    label_names = dict(enumerate(target_names))
    y_pred = pd.Series(y_pred, index=y_true.index).astype(y_true.dtype)
    preview = x_validation.loc[y_true.index].copy()
    preview["actual"] = y_true.map(label_names)
    preview["predicted"] = y_pred.map(label_names)
    preview["correct"] = preview["actual"] == preview["predicted"]

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
    from tabpfn import TabPFNClassifier

    start = time.perf_counter()
    if model_path is None:
        print("Loading the default TabPFN-3 classifier checkpoint...", flush=True)
    else:
        print(f"Loading the TabPFN checkpoint from {model_path}...", flush=True)

    classifier = TabPFNClassifier(
        model_path=model_path if model_path is not None else "auto",
        n_estimators=n_estimators,
        device=device,
        random_state=seed,
    )
    with windows_browser_auth_workaround():
        classifier.fit(x_train, y_train)
    probabilities = classifier.predict_proba(x_validation)
    if not np.isfinite(probabilities).all():
        raise RuntimeError("TabPFN returned non-finite probabilities.")
    predictions = classifier.classes_[np.argmax(probabilities, axis=1)]
    elapsed = time.perf_counter() - start
    return predictions, elapsed


def run_xgboost(
    x_train: pd.DataFrame,
    x_validation: pd.DataFrame,
    y_train: pd.Series,
    seed: int,
) -> tuple[object, float]:
    classifier = XGBClassifier(
        objective="binary:logistic",
        n_estimators=100,
        max_depth=3,
        learning_rate=0.1,
        subsample=1.0,
        colsample_bytree=1.0,
        eval_metric="logloss",
        random_state=seed,
        n_jobs=1,
    )

    start = time.perf_counter()
    classifier.fit(x_train, y_train)
    predictions = classifier.predict(x_validation)
    elapsed = time.perf_counter() - start
    return predictions, elapsed


def main() -> None:
    args = parse_args()
    x, y, target_names = load_breast_cancer_frame()
    x_train, x_validation, y_train, y_validation = train_test_split(
        x,
        y,
        test_size=args.validation_size,
        random_state=args.seed,
        stratify=y,
    )

    print("Dataset")
    print("=======")
    print(f"Rows:       {len(x)}")
    print(f"Features:   {x.shape[1]}")
    print(f"Training:   {len(x_train)}")
    print(f"Validation: {len(x_validation)}")
    print(f"Classes:    {', '.join(target_names)}")

    if not args.skip_tabpfn:
        tabpfn_model_path = resolve_model_path(
            args.tabpfn_model_path,
            CLASSIFIER_MODEL_FILENAME,
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
            target_names,
            tabpfn_elapsed,
        )
        print_prediction_preview(
            "TabPFN zero-shot",
            x_validation,
            y_validation,
            tabpfn_predictions,
            target_names,
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
        target_names,
        xgboost_elapsed,
    )
    print_prediction_preview(
        "XGBoost",
        x_validation,
        y_validation,
        xgboost_predictions,
        target_names,
        args.show_predictions,
    )


if __name__ == "__main__":
    main()
