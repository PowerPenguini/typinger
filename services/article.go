package services

import (
	"context"
	"fmt"

	"github.com/GlitchyGlitch/typinger/errs"
	"github.com/GlitchyGlitch/typinger/models"
	"github.com/go-pg/pg"
)

type ArticleRepo struct {
	DB *pg.DB
}

func (a *ArticleRepo) GetArticles(ctx context.Context, filter *models.ArticleFilter, first, offset int) ([]*models.Article, error) {
	var articles []*models.Article

	query := a.DB.Model(&articles).Order("id")

	if filter != nil {
		if filter.Title != nil {
			query.Where("title ILIKE ?", fmt.Sprintf("%%%s%%", *filter.Title))
		}
	}
	if first != 0 {
		query.Limit(first)
	}
	if offset != 0 {
		query.Offset(offset)
	}

	err := query.Select()
	if err != nil {
		return nil, errs.Internal(ctx)
	}
	if len(articles) == 0 {
		return nil, nil
	}

	return articles, nil
}

func (a *ArticleRepo) GetArticlesByUserIDs(ids []string) ([][]*models.Article, []error) { //TODO: move it to new loader to get context here
	var articles []*models.Article
	result := make([][]*models.Article, len(ids))
	aMap := make(map[string][]*models.Article, len(ids))

	err := a.DB.Model(&articles).Where("author in (?)", pg.In(ids)).Order("author").Select()

	if err != nil {
		return nil, []error{} // TODO: error hereDelete()
	}
	if len(articles) == 0 {
		return result, []error{} //handle errs not found here
	}
	for _, article := range articles {
		aMap[article.Author] = append(aMap[article.Author], article)
	}

	for i, id := range ids {
		result[i] = aMap[id]
	}
	return result, nil
}

func (a *ArticleRepo) CreateArticle(ctx context.Context, user *models.User, input *models.NewArticle) (*models.Article, error) {
	article := &models.Article{Title: input.Title, Content: input.Content, ThumbnailURL: input.ThumbnailURL, Author: user.ID}
	res, err := a.DB.Model(article).OnConflict("DO NOTHING").Insert()
	if err != nil {
		return nil, errs.Internal(ctx)
	}
	if res.RowsAffected() <= 0 {
		return nil, errs.Exists(ctx)
	}
	return article, nil
}

func (a *ArticleRepo) DeleteArticle(ctx context.Context, id string) (bool, error) {
	article := models.Article{ID: id}
	res, err := a.DB.Model(&article).Where("id = ?", article.ID).Delete()
	if err != nil {
		return false, errs.Internal(ctx)
	}
	if res.RowsAffected() <= 0 {
		errs.Add(ctx, errs.NotFound(ctx))
		return false, nil
	}
	return true, nil
}
