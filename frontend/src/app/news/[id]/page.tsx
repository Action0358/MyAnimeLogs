import type React from 'react'
import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Container } from '@/components/layout'
import { AnnouncementDetail } from '@/components/announcements'
import { ANNOUNCEMENTS, getAnnouncementById } from '@/constants/announcements'

// 静的エクスポートのため、ビルド時に作ったID以外は受け付けない（404になる）
export const dynamicParams = false

export const generateStaticParams = () => ANNOUNCEMENTS.map((a) => ({ id: a.id }))

type Props = { params: { id: string } }

export const generateMetadata = ({ params }: Props): Metadata => {
  const announcement = getAnnouncementById(params.id)
  if (!announcement) return { title: 'お知らせ | MyAnimeLogs' }
  return {
    title: `${announcement.title} | お知らせ | MyAnimeLogs`,
    description: announcement.body.replace(/\s+/g, ' ').slice(0, 120),
  }
}

/**
 * お知らせの詳細（#1111）。データはリポジトリの定数なので全件を静的生成する。
 * 一覧と同じ共通コンテナ・760px読み幅列に置く。
 */
const NewsDetailPage: React.FC<Props> = ({ params }) => {
  const announcement = getAnnouncementById(params.id)
  if (!announcement) notFound()

  return (
    <Container size="xl" className="pb-16 pt-5 md:pb-24 md:pt-10">
      <AnnouncementDetail announcement={announcement} />
    </Container>
  )
}

export default NewsDetailPage
