#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "httpx",
# ]
# ///
"""
Benchmark script for caveman compression.

Harvests README.md files from GitHub, measures token counts before/after
compression, tracks compression time, and generates an efficacy report.
"""

import argparse
import asyncio
import json
import os
import sys
import tempfile
import time
from dataclasses import dataclass, asdict
from pathlib import Path
from typing import List, Optional

import httpx


@dataclass
class BenchmarkResult:
    """Results from benchmarking a single README file."""
    repo_name: str
    original_size_bytes: int
    original_tokens: int
    compressed_size_bytes: int
    compressed_tokens: int
    compression_time_seconds: float
    tokens_saved: int
    tokens_saved_percent: float
    compression_ratio: float
    success: bool
    error_message: Optional[str] = None


def find_caveman_binary() -> str:
    """Locate the caveman binary."""
    # Check in project bin directory first
    project_bin = Path(__file__).parent / "caveman"
    if project_bin.exists():
        return str(project_bin)
    
    # Check in PATH
    import subprocess
    result = subprocess.run(
        ["which", "caveman"],
        capture_output=True,
        text=True
    )
    if result.returncode == 0:
        return result.stdout.strip()
    
    raise FileNotFoundError(
        "caveman binary not found. Please build it first with: task build:caveman"
    )


def find_tokcount_binary() -> str:
    """Locate the tokcount binary."""
    # Check in project bin directory first
    project_bin = Path(__file__).parent / "tokcount"
    if project_bin.exists():
        return str(project_bin)
    
    # Check in PATH
    import subprocess
    result = subprocess.run(
        ["which", "tokcount"],
        capture_output=True,
        text=True
    )
    if result.returncode == 0:
        return result.stdout.strip()
    
    raise FileNotFoundError(
        "tokcount binary not found. Please build it first with: task build:tokcount"
    )


async def count_tokens(text: str, tokcount_bin: str) -> Optional[int]:
    """Count tokens in text using tokcount CLI with JSON output."""
    try:
        proc = await asyncio.create_subprocess_exec(
            tokcount_bin,
            "--json",
            stdin=asyncio.subprocess.PIPE,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        
        stdout, stderr = await asyncio.wait_for(
            proc.communicate(input=text.encode('utf-8')),
            timeout=30.0
        )
        
        if proc.returncode != 0:
            print(f"tokcount failed: {stderr.decode('utf-8')}", file=sys.stderr)
            return None
        
        # Parse JSON output
        try:
            data = json.loads(stdout.decode('utf-8'))
            # Look for first OpenAI model token count
            for provider_data in data:
                if provider_data.get('provider') == 'OpenAI':
                    return provider_data.get('tokens')
            return None
        except json.JSONDecodeError as e:
            print(f"Failed to parse tokcount JSON output: {e}", file=sys.stderr)
            return None
        
    except asyncio.TimeoutError:
        print("tokcount timed out", file=sys.stderr)
        return None
    except Exception as e:
        print(f"Error counting tokens: {e}", file=sys.stderr)
        return None


async def compress_with_caveman(text: str, caveman_bin: str, config_path: str) -> tuple[Optional[str], float]:
    """Compress text using caveman --encode. Returns (compressed_text, duration)."""
    start_time = time.time()
    
    try:
        proc = await asyncio.create_subprocess_exec(
            caveman_bin,
            "--encode",
            "--config", config_path,
            stdin=asyncio.subprocess.PIPE,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        
        stdout, stderr = await asyncio.wait_for(
            proc.communicate(input=text.encode('utf-8')),
            timeout=600.0
        )
        
        duration = time.time() - start_time
        
        if proc.returncode != 0:
            print(f"caveman encode failed: {stderr.decode('utf-8')}", file=sys.stderr)
            return None, duration
        
        return stdout.decode('utf-8'), duration
    except asyncio.TimeoutError:
        duration = time.time() - start_time
        print("caveman encode timed out", file=sys.stderr)
        return None, duration
    except Exception as e:
        duration = time.time() - start_time
        print(f"Error compressing with caveman: {e}", file=sys.stderr)
        return None, duration


async def fetch_readme_from_github(repo_full_name: str, client: httpx.AsyncClient) -> Optional[str]:
    """Fetch README.md content from a GitHub repository."""
    # Try common README locations
    readme_paths = ["README.md", "readme.md", "Readme.md", "README.MD"]
    
    for readme_path in readme_paths:
        for branch in ["main", "master"]:
            url = f"https://raw.githubusercontent.com/{repo_full_name}/{branch}/{readme_path}"
            
            try:
                response = await client.get(url, timeout=10.0)
                if response.status_code == 200:
                    content = response.text
                    if content.strip():
                        return content
            except Exception:
                continue
    
    return None


async def search_github_repos(client: httpx.AsyncClient, count: int = 30) -> List[str]:
    """Search GitHub for popular repositories with README files."""
    repos = []
    per_page = 10
    
    # Search queries to get diverse repos
    queries = [
        "stars:>10000 language:python",
        "stars:>10000 language:javascript",
        "stars:>10000 language:go",
        "stars:>5000 language:rust",
        "stars:>5000 language:typescript",
    ]
    
    headers = {
        'Accept': 'application/vnd.github.v3+json'
    }
    
    # Add GitHub token if available
    github_token = os.environ.get('GITHUB_TOKEN')
    if github_token:
        headers['Authorization'] = f'token {github_token}'
    
    for query in queries:
        if len(repos) >= count:
            break
            
        url = f"https://api.github.com/search/repositories?q={query}&sort=stars&order=desc&per_page={per_page}"
        
        try:
            response = await client.get(url, headers=headers, timeout=10.0)
            response.raise_for_status()
            data = response.json()
            
            for item in data.get('items', []):
                if len(repos) >= count:
                    break
                repos.append(item['full_name'])
        
        except Exception as e:
            print(f"Error searching GitHub: {e}", file=sys.stderr)
            continue
    
    return repos[:count]


async def benchmark_readme(repo_name: str, caveman_bin: str, tokcount_bin: str, client: httpx.AsyncClient) -> BenchmarkResult:
    """Benchmark compression for a single README file."""
    print(f"Benchmarking {repo_name}...", file=sys.stderr)
    
    # Create temporary config with temporary cache directory
    temp_cache_dir = tempfile.mkdtemp(prefix="caveman-benchmark-")
    temp_cache_path = os.path.join(temp_cache_dir, "cache.db")
    
    # Create temporary config file
    config_data = {
        "model": os.environ.get("CAVEMAN_MODEL", ""),
        "base_url": os.environ.get("CAVEMAN_BASE_URL", ""),
        "api_key": os.environ.get("CAVEMAN_API_KEY", ""),
        "cache_path": temp_cache_path
    }
    
    with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as config_file:
        json.dump(config_data, config_file)
        config_path = config_file.name
    
    try:
        # Fetch README
        readme_content = await fetch_readme_from_github(repo_name, client)
        if not readme_content:
            return BenchmarkResult(
                repo_name=repo_name,
                original_size_bytes=0,
                original_tokens=0,
                compressed_size_bytes=0,
                compressed_tokens=0,
                compression_time_seconds=0.0,
                tokens_saved=0,
                tokens_saved_percent=0.0,
                compression_ratio=0.0,
                success=False,
                error_message="Failed to fetch README"
            )
        
        original_size = len(readme_content.encode('utf-8'))
        
        # Count original tokens
        original_tokens = await count_tokens(readme_content, tokcount_bin)
        if original_tokens is None:
            return BenchmarkResult(
                repo_name=repo_name,
                original_size_bytes=original_size,
                original_tokens=0,
                compressed_size_bytes=0,
                compressed_tokens=0,
                compression_time_seconds=0.0,
                tokens_saved=0,
                tokens_saved_percent=0.0,
                compression_ratio=0.0,
                success=False,
                error_message="Failed to count original tokens"
            )
        
        # Compress with caveman
        compressed_content, compression_time = await compress_with_caveman(readme_content, caveman_bin, config_path)
        if compressed_content is None:
            return BenchmarkResult(
                repo_name=repo_name,
                original_size_bytes=original_size,
                original_tokens=original_tokens,
                compressed_size_bytes=0,
                compressed_tokens=0,
                compression_time_seconds=compression_time,
                tokens_saved=0,
                tokens_saved_percent=0.0,
                compression_ratio=0.0,
                success=False,
                error_message="Failed to compress with caveman"
            )
        
        compressed_size = len(compressed_content.encode('utf-8'))
        
        # Count compressed tokens
        compressed_tokens = await count_tokens(compressed_content, tokcount_bin)
        if compressed_tokens is None:
            return BenchmarkResult(
                repo_name=repo_name,
                original_size_bytes=original_size,
                original_tokens=original_tokens,
                compressed_size_bytes=compressed_size,
                compressed_tokens=0,
                compression_time_seconds=compression_time,
                tokens_saved=0,
                tokens_saved_percent=0.0,
                compression_ratio=0.0,
                success=False,
                error_message="Failed to count compressed tokens"
            )
        
        # Calculate metrics
        tokens_saved = original_tokens - compressed_tokens
        tokens_saved_percent = (tokens_saved / original_tokens * 100) if original_tokens > 0 else 0.0
        compression_ratio = (compressed_size / original_size) if original_size > 0 else 0.0
        
        return BenchmarkResult(
            repo_name=repo_name,
            original_size_bytes=original_size,
            original_tokens=original_tokens,
            compressed_size_bytes=compressed_size,
            compressed_tokens=compressed_tokens,
            compression_time_seconds=compression_time,
            tokens_saved=tokens_saved,
            tokens_saved_percent=tokens_saved_percent,
            compression_ratio=compression_ratio,
            success=True
        )
    finally:
        # Clean up temporary files
        try:
            os.unlink(config_path)
        except Exception:
            pass
        try:
            import shutil
            shutil.rmtree(temp_cache_dir, ignore_errors=True)
        except Exception:
            pass


def generate_report(results: List[BenchmarkResult], output_file: str):
    """Generate a comprehensive benchmark report."""
    successful_results = [r for r in results if r.success]
    failed_results = [r for r in results if not r.success]
    
    if not successful_results:
        print("No successful benchmarks to report", file=sys.stderr)
        return
    
    # Calculate aggregate statistics
    total_original_tokens = sum(r.original_tokens for r in successful_results)
    total_compressed_tokens = sum(r.compressed_tokens for r in successful_results)
    total_tokens_saved = sum(r.tokens_saved for r in successful_results)
    avg_tokens_saved_percent = sum(r.tokens_saved_percent for r in successful_results) / len(successful_results)
    avg_compression_time = sum(r.compression_time_seconds for r in successful_results) / len(successful_results)
    avg_compression_ratio = sum(r.compression_ratio for r in successful_results) / len(successful_results)
    
    # Generate markdown report
    report = f"""# Caveman Compression Benchmark Report

## Summary

- **Total READMEs processed**: {len(results)}
- **Successful compressions**: {len(successful_results)}
- **Failed compressions**: {len(failed_results)}

## Aggregate Statistics

- **Total original tokens**: {total_original_tokens:,}
- **Total compressed tokens**: {total_compressed_tokens:,}
- **Total tokens saved**: {total_tokens_saved:,}
- **Average token reduction**: {avg_tokens_saved_percent:.2f}%
- **Average compression time**: {avg_compression_time:.3f} seconds
- **Average byte compression ratio**: {avg_compression_ratio:.3f}

## Efficacy Assessment

"""
    
    if avg_tokens_saved_percent >= 50:
        report += "✅ **Excellent**: Caveman compression achieves >50% token reduction on average.\n\n"
    elif avg_tokens_saved_percent >= 30:
        report += "✅ **Good**: Caveman compression achieves 30-50% token reduction on average.\n\n"
    elif avg_tokens_saved_percent >= 15:
        report += "⚠️ **Moderate**: Caveman compression achieves 15-30% token reduction on average.\n\n"
    else:
        report += "❌ **Poor**: Caveman compression achieves <15% token reduction on average.\n\n"
    
    report += f"""**Timing Performance**: Average compression time of {avg_compression_time:.3f}s per README suggests {"acceptable" if avg_compression_time < 5 else "slow"} performance for real-time use.

## Individual Results

| Repository | Original Tokens | Compressed Tokens | Tokens Saved | Reduction % | Time (s) |
|------------|-----------------|-------------------|--------------|-------------|----------|
"""
    
    for result in successful_results:
        report += f"| {result.repo_name} | {result.original_tokens:,} | {result.compressed_tokens:,} | {result.tokens_saved:,} | {result.tokens_saved_percent:.1f}% | {result.compression_time_seconds:.2f} |\n"
    
    if failed_results:
        report += "\n## Failed Compressions\n\n"
        for result in failed_results:
            report += f"- **{result.repo_name}**: {result.error_message}\n"
    
    report += "\n## Raw Data\n\n```json\n"
    report += json.dumps([asdict(r) for r in results], indent=2)
    report += "\n```\n"
    
    # Write report
    with open(output_file, 'w') as f:
        f.write(report)
    
    print(f"\nReport written to: {output_file}")
    print(f"\nSummary:")
    print(f"  Successful: {len(successful_results)}/{len(results)}")
    print(f"  Avg token reduction: {avg_tokens_saved_percent:.2f}%")
    print(f"  Avg compression time: {avg_compression_time:.3f}s")


async def main_async(serial: bool = False):
    """Main benchmark execution (async)."""
    print("Caveman Compression Benchmark", file=sys.stderr)
    print("=" * 50, file=sys.stderr)
    
    if serial:
        print("Running in SERIAL mode (one at a time)", file=sys.stderr)
    else:
        print("Running in PARALLEL mode (batches of 5)", file=sys.stderr)
    
    # Locate binaries
    try:
        caveman_bin = find_caveman_binary()
        tokcount_bin = find_tokcount_binary()
        print(f"Found caveman: {caveman_bin}", file=sys.stderr)
        print(f"Found tokcount: {tokcount_bin}", file=sys.stderr)
    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Create HTTP client
    async with httpx.AsyncClient() as client:
        # Search for repositories
        print("\nSearching for GitHub repositories...", file=sys.stderr)
        repos = await search_github_repos(client, count=30)
        
        if not repos:
            print("Error: No repositories found", file=sys.stderr)
            sys.exit(1)
        
        print(f"Found {len(repos)} repositories", file=sys.stderr)
        
        results = []
        
        if serial:
            # Run benchmarks serially (one at a time)
            print(f"\nProcessing {len(repos)} repositories serially...", file=sys.stderr)
            for idx, repo in enumerate(repos, 1):
                print(f"\n[{idx}/{len(repos)}] Processing {repo}...", file=sys.stderr)
                result = await benchmark_readme(repo, caveman_bin, tokcount_bin, client)
                results.append(result)
                
                if result.success:
                    print(f"  ✓ {result.repo_name}: {result.original_tokens:,} → {result.compressed_tokens:,} ({result.tokens_saved_percent:.1f}% saved, {result.compression_time_seconds:.2f}s)", file=sys.stderr)
                else:
                    print(f"  ✗ {result.repo_name}: {result.error_message}", file=sys.stderr)
        else:
            # Benchmark repositories concurrently (in batches to avoid overwhelming the system)
            batch_size = 5
            
            for i in range(0, len(repos), batch_size):
                batch = repos[i:i + batch_size]
                print(f"\nProcessing batch {i//batch_size + 1}/{(len(repos) + batch_size - 1)//batch_size}...", file=sys.stderr)
                
                # Run batch concurrently
                batch_tasks = [
                    benchmark_readme(repo, caveman_bin, tokcount_bin, client)
                    for repo in batch
                ]
                batch_results = await asyncio.gather(*batch_tasks)
                results.extend(batch_results)
                
                # Print batch results
                for result in batch_results:
                    if result.success:
                        print(f"  ✓ {result.repo_name}: {result.original_tokens:,} → {result.compressed_tokens:,} ({result.tokens_saved_percent:.1f}% saved, {result.compression_time_seconds:.2f}s)", file=sys.stderr)
                    else:
                        print(f"  ✗ {result.repo_name}: {result.error_message}", file=sys.stderr)
        
        # Generate report
        output_file = "caveman-benchmark-report.md"
        print(f"\nGenerating report...", file=sys.stderr)
        generate_report(results, output_file)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Benchmark caveman compression on GitHub README files"
    )
    parser.add_argument(
        "--serial",
        action="store_true",
        help="Run benchmarks serially (one at a time) instead of in parallel batches"
    )
    
    args = parser.parse_args()
    asyncio.run(main_async(serial=args.serial))


if __name__ == "__main__":
    main()