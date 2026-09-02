Code for blog post: https://blog.rasc.ch/2026/07/timesfm.html

The examples use TimesFM 3.0 and its `google/timesfm-3.0-pytorch` checkpoint.
The checkpoint is licensed for non-commercial, non-production use; review the
[model license](https://huggingface.co/google/timesfm-3.0-pytorch/blob/main/LICENSE)
before using it.

Download the 1.32 GB checkpoint with `task checkpoint`. The demos prefer the
local snapshot in `checkpoints/timesfm-3.0-pytorch` and fall back to downloading
it from Hugging Face when the local weights are absent.
