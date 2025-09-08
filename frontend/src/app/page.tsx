import type React from 'react';
import styles from './page.module.css';

const HomePage: React.FC = () => {
  return (
    <main className={styles.container}>
      <h1 className={styles.title}>
        Hello MyAnimeLogs!
      </h1>
    </main>
  );
};

export default HomePage;