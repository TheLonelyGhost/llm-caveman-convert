# Caveman Compression Benchmark Report

## Summary

- **Total READMEs processed**: 30
- **Successful compressions**: 30
- **Failed compressions**: 0

## Aggregate Statistics

- **Total original tokens**: 415,662
- **Total compressed tokens**: 227,867
- **Total tokens saved**: 187,795
- **Average token reduction**: 31.73%
- **Average compression time**: 67.455 seconds
- **Average byte compression ratio**: 0.670

## Efficacy Assessment

✅ **Good**: Caveman compression achieves 30-50% token reduction on average.

**Timing Performance**: Average compression time of 67.455s per README suggests slow performance for real-time use.

## Individual Results

| Repository | Original Tokens | Compressed Tokens | Tokens Saved | Reduction % | Time (s) |
|------------|-----------------|-------------------|--------------|-------------|----------|
| public-apis/public-apis | 59,135 | 52,157 | 6,978 | 11.8% | 486.70 |
| EbookFoundation/free-programming-books | 4,322 | 4,056 | 266 | 6.2% | 19.99 |
| donnemartin/system-design-primer | 26,479 | 23,351 | 3,128 | 11.8% | 192.67 |
| vinta/awesome-python | 20,118 | 18,948 | 1,170 | 5.8% | 157.89 |
| TheAlgorithms/Python | 795 | 739 | 56 | 7.0% | 16.66 |
| NousResearch/hermes-agent | 4,056 | 2,954 | 1,102 | 27.2% | 16.74 |
| Significant-Gravitas/AutoGPT | 2,850 | 520 | 2,330 | 81.8% | 4.91 |
| yt-dlp/yt-dlp | 40,814 | 19,049 | 21,765 | 53.3% | 165.39 |
| AUTOMATIC1111/stable-diffusion-webui | 3,362 | 755 | 2,607 | 77.5% | 6.39 |
| 521xueweihan/HelloGitHub | 2,103 | 1,933 | 170 | 8.1% | 29.70 |
| react/react | 1,210 | 1,030 | 180 | 14.9% | 16.45 |
| affaan-m/ECC | 22,266 | 20,362 | 1,904 | 8.6% | 170.81 |
| trekhleb/javascript-algorithms | 7,016 | 6,587 | 429 | 6.1% | 59.55 |
| Snailclimb/JavaGuide | 7,281 | 6,530 | 751 | 10.3% | 64.22 |
| airbnb/javascript | 35,375 | 33,226 | 2,149 | 6.1% | 278.05 |
| vercel/next.js | 6 | 6 | 0 | 0.0% | 1.06 |
| Chalarangelo/30-seconds-of-code | 372 | 313 | 59 | 15.9% | 9.53 |
| nodejs/node | 14,275 | 272 | 14,003 | 98.1% | 4.27 |
| mrdoob/three.js | 798 | 532 | 266 | 33.3% | 4.11 |
| axios/axios | 17,854 | 16,567 | 1,287 | 7.2% | 143.06 |
| avelino/awesome-go | 102,046 | 13 | 102,033 | 100.0% | 17.05 |
| ollama/ollama | 5,594 | 498 | 5,096 | 91.1% | 4.64 |
| golang/go | 311 | 611 | -300 | -96.5% | 5.58 |
| kubernetes/kubernetes | 971 | 456 | 515 | 53.0% | 4.26 |
| fatedier/frp | 11,921 | 10,603 | 1,318 | 11.1% | 90.50 |
| gin-gonic/gin | 3,227 | 3,074 | 153 | 4.7% | 31.68 |
| gohugoio/hugo | 4,312 | 1,181 | 3,131 | 72.6% | 8.17 |
| syncthing/syncthing | 1,003 | 597 | 406 | 40.5% | 5.13 |
| infiniflow/ragflow | 4,693 | 577 | 4,116 | 87.7% | 4.57 |
| junegunn/fzf | 11,097 | 370 | 10,727 | 96.7% | 3.94 |

## Raw Data

```json
[
  {
    "repo_name": "public-apis/public-apis",
    "original_size_bytes": 216753,
    "original_tokens": 59135,
    "compressed_size_bytes": 178360,
    "compressed_tokens": 52157,
    "compression_time_seconds": 486.70078706741333,
    "tokens_saved": 6978,
    "tokens_saved_percent": 11.800118373213833,
    "compression_ratio": 0.8228721171102591,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "EbookFoundation/free-programming-books",
    "original_size_bytes": 15921,
    "original_tokens": 4322,
    "compressed_size_bytes": 14849,
    "compressed_tokens": 4056,
    "compression_time_seconds": 19.986430168151855,
    "tokens_saved": 266,
    "tokens_saved_percent": 6.154558074965294,
    "compression_ratio": 0.9326675460084165,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "donnemartin/system-design-primer",
    "original_size_bytes": 109830,
    "original_tokens": 26479,
    "compressed_size_bytes": 95030,
    "compressed_tokens": 23351,
    "compression_time_seconds": 192.66939902305603,
    "tokens_saved": 3128,
    "tokens_saved_percent": 11.813134937119981,
    "compression_ratio": 0.8652462897204771,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "vinta/awesome-python",
    "original_size_bytes": 80364,
    "original_tokens": 20118,
    "compressed_size_bytes": 74262,
    "compressed_tokens": 18948,
    "compression_time_seconds": 157.89300107955933,
    "tokens_saved": 1170,
    "tokens_saved_percent": 5.815687444079928,
    "compression_ratio": 0.9240704793190981,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "TheAlgorithms/Python",
    "original_size_bytes": 2807,
    "original_tokens": 795,
    "compressed_size_bytes": 2527,
    "compressed_tokens": 739,
    "compression_time_seconds": 16.6569721698761,
    "tokens_saved": 56,
    "tokens_saved_percent": 7.044025157232705,
    "compression_ratio": 0.9002493765586035,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "NousResearch/hermes-agent",
    "original_size_bytes": 17642,
    "original_tokens": 4056,
    "compressed_size_bytes": 13090,
    "compressed_tokens": 2954,
    "compression_time_seconds": 16.736037969589233,
    "tokens_saved": 1102,
    "tokens_saved_percent": 27.169625246548325,
    "compression_ratio": 0.74197936741866,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "Significant-Gravitas/AutoGPT",
    "original_size_bytes": 11844,
    "original_tokens": 2850,
    "compressed_size_bytes": 2338,
    "compressed_tokens": 520,
    "compression_time_seconds": 4.906063079833984,
    "tokens_saved": 2330,
    "tokens_saved_percent": 81.75438596491227,
    "compression_ratio": 0.19739952718676124,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "yt-dlp/yt-dlp",
    "original_size_bytes": 179101,
    "original_tokens": 40814,
    "compressed_size_bytes": 68360,
    "compressed_tokens": 19049,
    "compression_time_seconds": 165.39180493354797,
    "tokens_saved": 21765,
    "tokens_saved_percent": 53.32728965551037,
    "compression_ratio": 0.38168407769917534,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "AUTOMATIC1111/stable-diffusion-webui",
    "original_size_bytes": 12924,
    "original_tokens": 3362,
    "compressed_size_bytes": 3028,
    "compressed_tokens": 755,
    "compression_time_seconds": 6.389850854873657,
    "tokens_saved": 2607,
    "tokens_saved_percent": 77.54312908982747,
    "compression_ratio": 0.23429278861033737,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "521xueweihan/HelloGitHub",
    "original_size_bytes": 6275,
    "original_tokens": 2103,
    "compressed_size_bytes": 6184,
    "compressed_tokens": 1933,
    "compression_time_seconds": 29.696529865264893,
    "tokens_saved": 170,
    "tokens_saved_percent": 8.083689966714218,
    "compression_ratio": 0.9854980079681275,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "react/react",
    "original_size_bytes": 5317,
    "original_tokens": 1210,
    "compressed_size_bytes": 4463,
    "compressed_tokens": 1030,
    "compression_time_seconds": 16.445487022399902,
    "tokens_saved": 180,
    "tokens_saved_percent": 14.87603305785124,
    "compression_ratio": 0.8393831107767538,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "affaan-m/ECC",
    "original_size_bytes": 90291,
    "original_tokens": 22266,
    "compressed_size_bytes": 80001,
    "compressed_tokens": 20362,
    "compression_time_seconds": 170.80622267723083,
    "tokens_saved": 1904,
    "tokens_saved_percent": 8.551154226174436,
    "compression_ratio": 0.8860351530052829,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "trekhleb/javascript-algorithms",
    "original_size_bytes": 25525,
    "original_tokens": 7016,
    "compressed_size_bytes": 23190,
    "compressed_tokens": 6587,
    "compression_time_seconds": 59.54793882369995,
    "tokens_saved": 429,
    "tokens_saved_percent": 6.114595210946408,
    "compression_ratio": 0.9085210577864838,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "Snailclimb/JavaGuide",
    "original_size_bytes": 25459,
    "original_tokens": 7281,
    "compressed_size_bytes": 22316,
    "compressed_tokens": 6530,
    "compression_time_seconds": 64.22137427330017,
    "tokens_saved": 751,
    "tokens_saved_percent": 10.314517236643319,
    "compression_ratio": 0.8765466043442398,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "airbnb/javascript",
    "original_size_bytes": 129525,
    "original_tokens": 35375,
    "compressed_size_bytes": 118502,
    "compressed_tokens": 33226,
    "compression_time_seconds": 278.04756784439087,
    "tokens_saved": 2149,
    "tokens_saved_percent": 6.074911660777385,
    "compression_ratio": 0.9148967380814514,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "vercel/next.js",
    "original_size_bytes": 23,
    "original_tokens": 6,
    "compressed_size_bytes": 23,
    "compressed_tokens": 6,
    "compression_time_seconds": 1.0580549240112305,
    "tokens_saved": 0,
    "tokens_saved_percent": 0.0,
    "compression_ratio": 1.0,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "Chalarangelo/30-seconds-of-code",
    "original_size_bytes": 1540,
    "original_tokens": 372,
    "compressed_size_bytes": 1259,
    "compressed_tokens": 313,
    "compression_time_seconds": 9.530431032180786,
    "tokens_saved": 59,
    "tokens_saved_percent": 15.86021505376344,
    "compression_ratio": 0.8175324675324676,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "nodejs/node",
    "original_size_bytes": 41467,
    "original_tokens": 14275,
    "compressed_size_bytes": 1209,
    "compressed_tokens": 272,
    "compression_time_seconds": 4.268367052078247,
    "tokens_saved": 14003,
    "tokens_saved_percent": 98.09457092819615,
    "compression_ratio": 0.02915571418236188,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "mrdoob/three.js",
    "original_size_bytes": 2969,
    "original_tokens": 798,
    "compressed_size_bytes": 2126,
    "compressed_tokens": 532,
    "compression_time_seconds": 4.112898111343384,
    "tokens_saved": 266,
    "tokens_saved_percent": 33.33333333333333,
    "compression_ratio": 0.7160660154934322,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "axios/axios",
    "original_size_bytes": 75661,
    "original_tokens": 17854,
    "compressed_size_bytes": 67298,
    "compressed_tokens": 16567,
    "compression_time_seconds": 143.05620408058167,
    "tokens_saved": 1287,
    "tokens_saved_percent": 7.208468690489527,
    "compression_ratio": 0.8894674931602807,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "avelino/awesome-go",
    "original_size_bytes": 394594,
    "original_tokens": 102046,
    "compressed_size_bytes": 65,
    "compressed_tokens": 13,
    "compression_time_seconds": 17.05235481262207,
    "tokens_saved": 102033,
    "tokens_saved_percent": 99.98726064715913,
    "compression_ratio": 0.00016472627561493587,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "ollama/ollama",
    "original_size_bytes": 19150,
    "original_tokens": 5594,
    "compressed_size_bytes": 1827,
    "compressed_tokens": 498,
    "compression_time_seconds": 4.6436591148376465,
    "tokens_saved": 5096,
    "tokens_saved_percent": 91.09760457633178,
    "compression_ratio": 0.09540469973890339,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "golang/go",
    "original_size_bytes": 1454,
    "original_tokens": 311,
    "compressed_size_bytes": 2642,
    "compressed_tokens": 611,
    "compression_time_seconds": 5.581117153167725,
    "tokens_saved": -300,
    "tokens_saved_percent": -96.46302250803859,
    "compression_ratio": 1.8170563961485557,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "kubernetes/kubernetes",
    "original_size_bytes": 4387,
    "original_tokens": 971,
    "compressed_size_bytes": 1991,
    "compressed_tokens": 456,
    "compression_time_seconds": 4.258912801742554,
    "tokens_saved": 515,
    "tokens_saved_percent": 53.03810504634397,
    "compression_ratio": 0.4538408935491224,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "fatedier/frp",
    "original_size_bytes": 44668,
    "original_tokens": 11921,
    "compressed_size_bytes": 37836,
    "compressed_tokens": 10603,
    "compression_time_seconds": 90.498850107193,
    "tokens_saved": 1318,
    "tokens_saved_percent": 11.05611945306602,
    "compression_ratio": 0.8470493418106922,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "gin-gonic/gin",
    "original_size_bytes": 11827,
    "original_tokens": 3227,
    "compressed_size_bytes": 10756,
    "compressed_tokens": 3074,
    "compression_time_seconds": 31.676017999649048,
    "tokens_saved": 153,
    "tokens_saved_percent": 4.7412457390765415,
    "compression_ratio": 0.909444491417942,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "gohugoio/hugo",
    "original_size_bytes": 13701,
    "original_tokens": 4312,
    "compressed_size_bytes": 4339,
    "compressed_tokens": 1181,
    "compression_time_seconds": 8.165589094161987,
    "tokens_saved": 3131,
    "tokens_saved_percent": 72.6113172541744,
    "compression_ratio": 0.3166922122472812,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "syncthing/syncthing",
    "original_size_bytes": 3999,
    "original_tokens": 1003,
    "compressed_size_bytes": 2506,
    "compressed_tokens": 597,
    "compression_time_seconds": 5.131266117095947,
    "tokens_saved": 406,
    "tokens_saved_percent": 40.47856430707876,
    "compression_ratio": 0.6266566641660415,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "infiniflow/ragflow",
    "original_size_bytes": 17021,
    "original_tokens": 4693,
    "compressed_size_bytes": 2345,
    "compressed_tokens": 577,
    "compression_time_seconds": 4.569870948791504,
    "tokens_saved": 4116,
    "tokens_saved_percent": 87.70509269124229,
    "compression_ratio": 0.13777098877856764,
    "success": true,
    "error_message": null
  },
  {
    "repo_name": "junegunn/fzf",
    "original_size_bytes": 41829,
    "original_tokens": 11097,
    "compressed_size_bytes": 1670,
    "compressed_tokens": 370,
    "compression_time_seconds": 3.942293882369995,
    "tokens_saved": 10727,
    "tokens_saved_percent": 96.6657655222132,
    "compression_ratio": 0.039924454325946115,
    "success": true,
    "error_message": null
  }
]
```
