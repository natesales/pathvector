import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useThemeContext from '@theme/hooks/useThemeContext';
import styles from './index.module.css';
import HomepageFeatures from '../components/HomepageFeatures';

function HomepageHeader() {
    const {isDarkTheme} = useThemeContext();
    return (
        <header style={{
            textAlign: "center"
        }}>
            <div className="container">
                <img width={"50%"} src={isDarkTheme ? "/img/full-white.svg" : "/img/full-black.svg"}
                     alt="Pathvector Logo"/>
                <div className={styles.buttons}>
                    <Link
                        className="button button--secondary button--lg"
                        to="/docs/about">
                        Learn More
                    </Link>
                </div>
            </div>
        </header>
    );
}

export default function Home() {
    return (
        <Layout
            title={`Pathvector | Edge Routing Platform`}
            description="Description will go into a meta tag in <head />">
            <HomepageHeader/>
            <main>
                <HomepageFeatures/>
            </main>
        </Layout>
    );
}
