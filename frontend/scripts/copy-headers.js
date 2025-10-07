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
  console.error('Error: Environment must be \"staging\" or \"production\". (エラー: 環境は \"staging\" または \"production\" である必要があります)');
  console.error('Usage: node copy-headers.js <staging|production> (使用方法: node copy-headers.js <staging|production>)');
  process.exit(1);
}

// ファイルパスを定義
const sourceFile = path.join(__dirname, '..', 'public', `_headers.${environment}`);
const destFile = path.join(__dirname, '..', 'out', '_headers');

// ソースファイルの存在確認
if (!fs.existsSync(sourceFile)) {
  console.error(`エラー: ソースファイルが見つかりません: ${sourceFile}`);
  process.exit(1);
}

// 出力ディレクトリの存在確認
const outDir = path.dirname(destFile);
if (!fs.existsSync(outDir)) {
  console.error(`エラー: 出力ディレクトリが見つかりません: ${outDir}`);
  console.error('出力ディレクトリを作成するために、まず "next build" を実行してください。');
  process.exit(1);
}

// ファイルをコピー
try {
  fs.copyFileSync(sourceFile, destFile);
  console.log(`✓ ${path.basename(sourceFile)} を ${destFile} にコピーしました`);
} catch (error) {
  console.error(`エラー: ファイルのコピーに失敗しました: ${error.message}`);
  process.exit(1);
}
