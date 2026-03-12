---
search: false
---

A GPU is not required to install Olares, but is necessary for most AI applications. Only NVIDIA GPUs are supported.

- **Architecture**: Turing or newer (GTX 16xx, RTX 20xx, 30xx, 40xx, 50xx series, and later).
:::info
GPUs with older architectures are not recognized by Olares. AI applications that require GPU access will not run.
:::
- **VRAM**: At least 8 GB recommended. Even on a supported GPU, limited VRAM will prevent many AI applications from running.

:::details Not sure if your GPU is supported?
Run the following command and check the code name prefix in the output:

```bash
lspci | grep -i nvidia
```

Example output:

```
3b:00.0 VGA compatible controller: NVIDIA Corporation AD102 [GeForce RTX 4090] (rev a1)
```

The code name prefix maps to GPU architecture as follows:

| Code name prefix | Architecture | Supported |
|:---|:---|:---:|
| GB | Blackwell | ✓ |
| AD | Ada Lovelace | ✓ |
| GA | Ampere | ✓ |
| TU | Turing | ✓ |
| GP | Pascal | ✗ |
| GM | Maxwell | ✗ |

You can also verify against the [full compatible GPU list](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus) from the NVIDIA open driver repository.
:::
