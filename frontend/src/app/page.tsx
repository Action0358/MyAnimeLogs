import Link from 'next/link'
import type React from 'react'
import styles from './page.module.css' // CSS Modulesを使用

// import Header from '@/components/Header';

const Top: React.FC = () => {
  return (
    <>
      {/* <Header /> ✅ 将来追加するヘッダー */}
      <main className={styles.container}>
        <h1 className={styles.title}>
          <span className={styles.red}>「名探偵コナン」</span>
          視聴記録アプリへようこそ！
        </h1>

        <div className={styles.grid}>
          {/* 左側：アプリ説明 */}
          <section className={`${styles.section} ${styles.descriptionSection}`}>
            <h2 className={styles.sectionTitle}>
              「名探偵コナン」視聴記録アプリとは...
            </h2>
            <ul className={styles.list}>
              <li>膨大な話数の中から視聴したエピソードを管理</li>
              <li>視聴進捗をグラフで可視化</li>
              <li>プレイリストで好きな話を管理</li>
              <li>映画予習用にも使える！</li>
            </ul>
          </section>

          {/* 右側：ログイン＋新規登録を縦に */}
          <div className={styles.authContainer}>
            <section className={styles.section}>
              <h2 className={styles.sectionTitle}>アカウントをお持ちの方</h2>
              <div className={styles.buttonWrapper}>
                <Link href="/login" className={styles.loginBtn}>
                  ログイン
                </Link>
              </div>
            </section>

            <section className={styles.section}>
              <h2 className={styles.sectionTitle}>
                アカウントをまだお持ちでない方
              </h2>
              <div className={styles.buttonWrapper}>
                <Link href="/register" className={styles.registerBtn}>
                  新規登録
                </Link>
              </div>
            </section>
          </div>
        </div>
      </main>
    </>
  )
}

export default Top
