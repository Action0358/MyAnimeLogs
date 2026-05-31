import { createClient } from '@supabase/supabase-js'

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!

// Supabaseクライアントが ?code= を消費し URL から削除する前にキャプチャ。
// PKCEフロー処理中のINITIAL_SESSION発火時にrecoveryセッションを同期しないための判定に使用する。
export const initialUrlHadCode =
  typeof window !== 'undefined'
    ? new URLSearchParams(window.location.search).has('code')
    : false

export const supabase = createClient(supabaseUrl, supabaseAnonKey)
