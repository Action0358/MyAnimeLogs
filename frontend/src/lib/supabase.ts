import { createClient } from '@supabase/supabase-js'

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!

// Supabaseクライアントが ?code= を消費し URL から削除する前にキャプチャ。
// PKCEフロー処理中のINITIAL_SESSION発火時にrecoveryセッションを同期しないための判定に使用する。
// 消費後に false へクリアできるよう mutable オブジェクトでエクスポートし、
// React Strict Mode のダブルマウントや useEffect 再実行時に値が残り続けるのを防ぐ。
export const initialUrlHadCode = {
  value: typeof window !== 'undefined' && new URLSearchParams(window.location.search).has('code'),
}

export const supabase = createClient(supabaseUrl, supabaseAnonKey)
