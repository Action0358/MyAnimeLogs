import { createClient } from '@supabase/supabase-js'

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!

// /auth/callback が ?code= を交換して URL から消す前にキャプチャ。
// PKCEフロー処理中のINITIAL_SESSION発火時に古いセッションを同期しないための判定に使用する。
// 消費後に false へクリアできるよう mutable オブジェクトでエクスポートし、
// React Strict Mode のダブルマウントや useEffect 再実行時に値が残り続けるのを防ぐ。
export const initialUrlHadCode = {
  value: typeof window !== 'undefined' && new URLSearchParams(window.location.search).has('code'),
}

export const supabase = createClient(supabaseUrl, supabaseAnonKey, {
  auth: {
    // implicit フローは URL のフラグメントにトークンが載り、履歴などに残りやすい
    flowType: 'pkce',
    // コードの交換は /auth/callback だけで行う。自動検出を有効のままにすると、起動時に
    // クライアントが先に交換して code verifier を消すため、callback の交換が必ず失敗する
    detectSessionInUrl: false,
  },
})
