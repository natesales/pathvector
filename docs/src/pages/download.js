import React from 'react';
import Layout from '@theme/Layout';
import {DownloadForm} from "../components/DownloadForm";

export default function Home() {
    return (
        <Layout
            title={`Download`}
            description="Pathvector is a declarative edge routing platform that automates route optimization and control plane configuration with secure and repeatable routing policy.">
            <main>
                <section>
                    <DownloadForm/>
                </section>
            </main>
        </Layout>
    );
}
