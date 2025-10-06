#!/usr/bin/env node

/**
 * クロスプラットフォーム対応のヘッダーファイルコピースクリプト
 * 環境に応じた _headers ファイルを out/ ディレクトリにコピーします
 */

const fs = require('fs');
const path = require('path');

// コマンドライン引数から環境を取得
const environment = process.argv[2];

if (!environment || !['staging', 'production'].includes(environment)) {
  console.error('Error: Environment must be "staging" or "production"');
  console.error('Usage: node copy-headers.js <staging|production>');
  process.exit(1);
}

// ファイルパスを定義
const sourceFile = path.join(__dirname, '..', 'public', `_headers.${environment}`);
const destFile = path.join(__dirname, '..', 'out', '_headers');

// ソースファイルの存在確認
if (!fs.existsSync(sourceFile)) {
  console.error(`Error: Source file not found: ${sourceFile}`);
  process.exit(1);
}

// 出力ディレクトリの存在確認
const outDir = path.dirname(destFile);
if (!fs.existsSync(outDir)) {
  console.error(`Error: Output directory not found: ${outDir}`);
  console.error('Please run "next build" first to create the output directory.');
  process.exit(1);
}

// ファイルをコピー
try {
  fs.copyFileSync(sourceFile, destFile);
  console.log(`✓ Copied ${path.basename(sourceFile)} to ${destFile}`);
} catch (error) {
  console.error(`Error: Failed to copy file: ${error.message}`);
  process.exit(1);
}
